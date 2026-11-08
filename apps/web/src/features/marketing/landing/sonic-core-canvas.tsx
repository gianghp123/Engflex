import { Canvas, useFrame } from "@react-three/fiber";
import { useMemo, useRef } from "react";
import * as THREE from "three";

export interface SonicCoreTelemetry {
	/** Normalized page scroll 0..1 */
	scrollProgress: number;
	/** Active narrative chapter 0..5 */
	activeChapter: number;
	/** Selected CEFR tier 0 (A1) .. 5 (C2) */
	activeTierIndex: number;
	/** Decaying voice/acoustic pulse intensity 0..1 */
	voiceEnergy: number;
	/** Normalized pointer coordinates -1..1 */
	pointerX: number;
	pointerY: number;
	/** Pointer speed 0..1 */
	pointerVelocity: number;
	/** User prefers-reduced-motion */
	reducedMotion: boolean;
}

interface SonicCoreCanvasProps {
	telemetryRef: React.RefObject<SonicCoreTelemetry>;
}

const VERTEX_SHADER = /* glsl */ `
  uniform float uTime;
  uniform float uScroll;
  uniform float uVoiceEnergy;
  uniform float uPointerVel;
  uniform vec2 uPointer;
  uniform float uMorphMode;

  varying vec3 vNormal;
  varying vec3 vPosition;
  varying float vDisplacement;
  varying float vFresnel;

  // Multi-octave harmonic interference for fluid organic acoustic waves
  float waveField(vec3 p, float t, float freqBoost) {
    float f = 2.1 + freqBoost * 1.6;
    float w1 = sin(p.x * f + t * 1.5) * cos(p.y * f - t * 1.2) * sin(p.z * f + t * 0.9);
    float w2 = sin(p.y * (f * 1.85) - t * 2.1) * cos(p.z * (f * 1.85) + t * 1.7) * 0.5;
    float w3 = cos(length(p.xy) * (4.0 + uVoiceEnergy * 5.0) - t * 3.4) * (0.22 + uVoiceEnergy * 0.45);
    return w1 + w2 + w3;
  }

  void main() {
    vNormal = normalize(normalMatrix * normal);

    vec3 pos = position;
    float pointerInfluence = clamp(1.0 - length(pos.xy - uPointer * 1.4) * 0.55, 0.0, 1.0);
    float freqBoost = uMorphMode * 0.35 + uVoiceEnergy * 0.85 + uPointerVel * 0.5;

    float n = waveField(normal, uTime, freqBoost);
    float amp = 0.16 + uVoiceEnergy * 0.26 + uPointerVel * 0.14 + pointerInfluence * 0.12;
    float displacement = n * amp;

    vec3 displacedPosition = pos + normal * displacement;
    vPosition = displacedPosition;
    vDisplacement = displacement;

    vec4 mvPosition = modelViewMatrix * vec4(displacedPosition, 1.0);
    vec3 viewDir = normalize(-mvPosition.xyz);
    vFresnel = pow(1.0 - abs(dot(vNormal, viewDir)), 2.4);

    gl_Position = projectionMatrix * mvPosition;
  }
`;

const FRAGMENT_SHADER = /* glsl */ `
  uniform float uTime;
  uniform float uVoiceEnergy;
  uniform vec3 uColorPrimary;
  uniform vec3 uColorSecondary;
  uniform vec3 uColorAccent;
  uniform vec3 uColorDeep;

  varying vec3 vNormal;
  varying vec3 vPosition;
  varying float vDisplacement;
  varying float vFresnel;

  void main() {
    vec3 lightDir1 = normalize(vec3(1.2, 1.6, 2.0));
    vec3 lightDir2 = normalize(vec3(-1.5, -1.0, 1.2));

    float diff1 = max(dot(vNormal, lightDir1), 0.0);
    float diff2 = max(dot(vNormal, lightDir2), 0.0);

    float waveMix = smoothstep(-0.28, 0.32, vDisplacement);
    vec3 baseCol = mix(uColorDeep, uColorPrimary, 0.35 + diff1 * 0.65);
    baseCol = mix(baseCol, uColorSecondary, waveMix * (0.55 + uVoiceEnergy * 0.35));
    baseCol = mix(baseCol, uColorAccent, diff2 * 0.42);

    // Iridescent Fresnel rim + specular acoustic crest
    vec3 rimCol = mix(uColorAccent, uColorSecondary, 0.5 + 0.5 * sin(uTime * 1.4 + vPosition.y * 2.0));
    vec3 finalColor = baseCol + rimCol * vFresnel * (0.85 + uVoiceEnergy * 0.6);

    // Subtle harmonic scanlines on high voice energy
    float scan = sin(vPosition.y * 24.0 - uTime * 6.0) * 0.5 + 0.5;
    finalColor += uColorSecondary * scan * uVoiceEnergy * 0.18;

    gl_FragColor = vec4(finalColor, 0.96);
  }
`;

const TIER_COLORS: readonly [string, string][] = [
	["#ff9e3d", "#6c4df0"], // A1: Apricot + Violet
	["#ffb347", "#7c5dfa"], // A2: Warm Amber + Violet
	["#6c4df0", "#3d9bff"], // B1: Electric Violet + Sky
	["#8a73ff", "#ff9e3d"], // B2: Luminous Violet + Apricot
	["#3d9bff", "#6c4df0"], // C1: Sonic Sky + Violet
	["#5eead4", "#8a73ff"], // C2: Cyber Teal/Sky + Violet
];

function AcousticScene({ telemetryRef }: SonicCoreCanvasProps) {
	"use no memo";

	const rootGroupRef = useRef<THREE.Group>(null);
	const coreMeshRef = useRef<THREE.Mesh>(null);
	const wireShellRef = useRef<THREE.Mesh>(null);
	const orbitsGroupRef = useRef<THREE.Group>(null);
	const particlesRef = useRef<THREE.Points>(null);
	const pulseRingRef = useRef<THREE.Mesh>(null);

	const uniforms = useMemo(
		() => ({
			uTime: { value: 0 },
			uScroll: { value: 0 },
			uVoiceEnergy: { value: 0 },
			uPointerVel: { value: 0 },
			uPointer: { value: new THREE.Vector2(0, 0) },
			uMorphMode: { value: 0 },
			uColorDeep: { value: new THREE.Color("#120b2e") },
			uColorPrimary: { value: new THREE.Color("#6c4df0") },
			uColorSecondary: { value: new THREE.Color("#ff9e3d") },
			uColorAccent: { value: new THREE.Color("#3d9bff") },
		}),
		[],
	);

	const targetPrimary = useMemo(() => new THREE.Color("#6c4df0"), []);
	const targetSecondary = useMemo(() => new THREE.Color("#ff9e3d"), []);
	const targetAccent = useMemo(() => new THREE.Color("#3d9bff"), []);

	// Build 720 orbital particles once
	const particleGeometry = useMemo(() => {
		const count = 720;
		const positions = new Float32Array(count * 3);
		for (let i = 0; i < count; i++) {
			const radius = 2.2 + ((i * 19) % 100) * 0.042;
			const theta = i * 0.37;
			const phi = ((i * 73) % 180) * (Math.PI / 180);
			positions[i * 3] = radius * Math.sin(phi) * Math.cos(theta);
			positions[i * 3 + 1] = radius * Math.cos(phi) * 0.65;
			positions[i * 3 + 2] = radius * Math.sin(phi) * Math.sin(theta);
		}
		const geom = new THREE.BufferGeometry();
		geom.setAttribute("position", new THREE.BufferAttribute(positions, 3));
		return geom;
	}, []);

	// 6 concentric CEFR rings config
	const cefrRings = useMemo(
		() =>
			[0, 1, 2, 3, 4, 5].map((idx) => ({
				id: `cefr-ring-${idx}`,
				idx,
				radius: 2.15 + idx * 0.28,
				tiltX: 0.45 + idx * 0.18,
				tiltY: idx * 0.52,
				speed: (idx % 2 === 0 ? 1 : -1) * (0.28 - idx * 0.03),
				color: idx < 2 ? "#ff9e3d" : idx < 4 ? "#8a73ff" : "#3d9bff",
			})),
		[],
	);

	useFrame((state, delta) => {
		const t = telemetryRef.current;
		if (!t) return;

		const dt = Math.min(delta, 0.05);
		const speedFactor = t.reducedMotion ? 0.08 : 1.0;
		uniforms.uTime.value += dt * speedFactor;

		// Decay voice energy smoothly
		if (t.voiceEnergy > 0.001) {
			t.voiceEnergy = Math.max(0, t.voiceEnergy - dt * 0.65);
		}
		if (t.pointerVelocity > 0.001) {
			t.pointerVelocity = Math.max(0, t.pointerVelocity - dt * 1.8);
		}

		uniforms.uScroll.value = THREE.MathUtils.lerp(
			uniforms.uScroll.value,
			t.scrollProgress,
			0.08,
		);
		uniforms.uVoiceEnergy.value = THREE.MathUtils.lerp(
			uniforms.uVoiceEnergy.value,
			t.voiceEnergy,
			0.14,
		);
		uniforms.uPointerVel.value = THREE.MathUtils.lerp(
			uniforms.uPointerVel.value,
			t.pointerVelocity,
			0.12,
		);
		uniforms.uPointer.value.x = THREE.MathUtils.lerp(
			uniforms.uPointer.value.x,
			t.pointerX,
			0.09,
		);
		uniforms.uPointer.value.y = THREE.MathUtils.lerp(
			uniforms.uPointer.value.y,
			t.pointerY,
			0.09,
		);
		uniforms.uMorphMode.value = THREE.MathUtils.lerp(
			uniforms.uMorphMode.value,
			t.activeChapter,
			0.06,
		);

		// Determine palette by active chapter & CEFR tier
		const tierPair = TIER_COLORS[t.activeTierIndex] ?? TIER_COLORS[0];
		if (t.activeChapter === 1) {
			targetSecondary.set(tierPair[0]);
			targetPrimary.set(tierPair[1]);
			targetAccent.set("#3d9bff");
		} else if (t.activeChapter === 2) {
			targetPrimary.set("#6c4df0");
			targetSecondary.set("#ff9e3d");
			targetAccent.set("#ff4d6d");
		} else if (t.activeChapter === 3) {
			targetPrimary.set("#3d9bff");
			targetSecondary.set("#8a73ff");
			targetAccent.set("#ff9e3d");
		} else if (t.activeChapter === 4) {
			targetPrimary.set("#5eead4");
			targetSecondary.set("#6c4df0");
			targetAccent.set("#3d9bff");
		} else {
			targetPrimary.set("#6c4df0");
			targetSecondary.set("#ff9e3d");
			targetAccent.set("#3d9bff");
		}

		uniforms.uColorPrimary.value.lerp(targetPrimary, 0.06);
		uniforms.uColorSecondary.value.lerp(targetSecondary, 0.06);
		uniforms.uColorAccent.value.lerp(targetAccent, 0.06);

		// Position the 3D stage responsively across scroll chapters
		const isWide = state.viewport.width > 6.5;
		let targetX = 0;
		let targetY = 0;
		let targetScale = isWide ? 1.0 : 0.76;

		if (isWide) {
			if (t.activeChapter === 0) {
				targetX = 1.65;
				targetY = 0.15;
				targetScale = 1.02;
			} else if (t.activeChapter === 1) {
				targetX = -1.75;
				targetY = -0.05;
				targetScale = 0.96;
			} else if (t.activeChapter === 2) {
				targetX = 1.6;
				targetY = 0.0;
				targetScale = 0.94;
			} else if (t.activeChapter === 3) {
				targetX = 0.0;
				targetY = 0.1;
				targetScale = 0.85;
			} else if (t.activeChapter === 4) {
				targetX = -1.55;
				targetY = 0.0;
				targetScale = 0.92;
			} else {
				targetX = 0.0;
				targetY = 0.25;
				targetScale = 1.1;
			}
		}

		if (rootGroupRef.current) {
			rootGroupRef.current.position.x = THREE.MathUtils.lerp(
				rootGroupRef.current.position.x,
				targetX + uniforms.uPointer.value.x * 0.28,
				0.06,
			);
			rootGroupRef.current.position.y = THREE.MathUtils.lerp(
				rootGroupRef.current.position.y,
				targetY + uniforms.uPointer.value.y * 0.22,
				0.06,
			);
			const nextScale = THREE.MathUtils.lerp(
				rootGroupRef.current.scale.x,
				targetScale + uniforms.uVoiceEnergy.value * 0.08,
				0.08,
			);
			rootGroupRef.current.scale.setScalar(nextScale);

			rootGroupRef.current.rotation.y = THREE.MathUtils.lerp(
				rootGroupRef.current.rotation.y,
				uniforms.uPointer.value.x * 0.42 +
					uniforms.uScroll.value * Math.PI * 1.4,
				0.06,
			);
			rootGroupRef.current.rotation.x = THREE.MathUtils.lerp(
				rootGroupRef.current.rotation.x,
				-uniforms.uPointer.value.y * 0.32 + t.activeTierIndex * 0.12,
				0.06,
			);
		}

		if (coreMeshRef.current) {
			coreMeshRef.current.rotation.y += dt * 0.22 * speedFactor;
			coreMeshRef.current.rotation.z += dt * 0.14 * speedFactor;
		}

		if (wireShellRef.current) {
			wireShellRef.current.rotation.y -= dt * 0.16 * speedFactor;
			wireShellRef.current.rotation.x += dt * 0.11 * speedFactor;
			const shellScale =
				1.18 +
				Math.sin(uniforms.uTime.value * 1.6) * 0.03 +
				uniforms.uVoiceEnergy.value * 0.18;
			wireShellRef.current.scale.setScalar(shellScale);
		}

		if (orbitsGroupRef.current) {
			const children = orbitsGroupRef.current.children;
			for (let i = 0; i < children.length; i++) {
				const ringGroup = children[i];
				const cfg = cefrRings[i];
				if (!ringGroup || !cfg) continue;
				ringGroup.rotation.z += dt * cfg.speed * speedFactor;
				const isSelected = i === t.activeTierIndex;
				const targetRingScale = isSelected ? 1.07 : 1.0;
				ringGroup.scale.setScalar(
					THREE.MathUtils.lerp(ringGroup.scale.x, targetRingScale, 0.1),
				);
			}
		}

		if (particlesRef.current) {
			particlesRef.current.rotation.y += dt * 0.07 * speedFactor;
			particlesRef.current.rotation.x =
				Math.sin(uniforms.uTime.value * 0.3) * 0.15;
		}

		if (pulseRingRef.current) {
			const pulse =
				1.3 +
				((uniforms.uTime.value * 0.65) % 1) * 2.2 +
				uniforms.uVoiceEnergy.value * 0.9;
			pulseRingRef.current.scale.setScalar(pulse);
			const mat = pulseRingRef.current.material as THREE.MeshBasicMaterial;
			if (mat) {
				mat.opacity = Math.max(
					0,
					(1.0 - ((uniforms.uTime.value * 0.65) % 1)) *
						(0.22 + uniforms.uVoiceEnergy.value * 0.45),
				);
			}
		}
	});

	return (
		<group ref={rootGroupRef}>
			{/* Ambient & directional studio lighting */}
			<ambientLight intensity={0.65} />
			<pointLight position={[4, 4, 4]} intensity={2.4} color="#8a73ff" />
			<pointLight position={[-4, -3, 3]} intensity={1.8} color="#ff9e3d" />
			<pointLight position={[0, 3, -4]} intensity={1.5} color="#3d9bff" />

			{/* Central Acoustic Fluid Core */}
			<mesh ref={coreMeshRef}>
				<icosahedronGeometry args={[1.42, 48]} />
				<shaderMaterial
					vertexShader={VERTEX_SHADER}
					fragmentShader={FRAGMENT_SHADER}
					uniforms={uniforms}
					transparent
				/>
			</mesh>

			{/* Geodesic Acoustic Cage */}
			<mesh ref={wireShellRef}>
				<icosahedronGeometry args={[1.48, 2]} />
				<meshBasicMaterial
					color="#a594ff"
					wireframe
					transparent
					opacity={0.16}
				/>
			</mesh>

			{/* Expanding Sonic Wave Ring */}
			<mesh ref={pulseRingRef} rotation={[Math.PI / 3, 0.2, 0]}>
				<torusGeometry args={[1.2, 0.012, 16, 96]} />
				<meshBasicMaterial
					color="#ff9e3d"
					transparent
					opacity={0.28}
					blending={THREE.AdditiveBlending}
				/>
			</mesh>

			{/* 6 CEFR Orbital Rings & Nodes (A1 -> C2) */}
			<group ref={orbitsGroupRef}>
				{cefrRings.map((ring) => (
					<group key={ring.id} rotation={[ring.tiltX, ring.tiltY, 0]}>
						<mesh>
							<torusGeometry args={[ring.radius, 0.009, 12, 110]} />
							<meshBasicMaterial
								color={ring.color}
								transparent
								opacity={0.28}
								blending={THREE.AdditiveBlending}
							/>
						</mesh>
						{/* Satellite Node on the orbit */}
						<mesh position={[ring.radius, 0, 0]}>
							<sphereGeometry args={[0.075, 16, 16]} />
							<meshBasicMaterial color={ring.color} />
						</mesh>
					</group>
				))}
			</group>

			{/* GPU Orbital Particle Field */}
			<points ref={particlesRef} geometry={particleGeometry}>
				<pointsMaterial
					size={0.032}
					color="#c9beff"
					transparent
					opacity={0.55}
					blending={THREE.AdditiveBlending}
					depthWrite={false}
				/>
			</points>
		</group>
	);
}

export default function SonicCoreCanvas({
	telemetryRef,
}: SonicCoreCanvasProps) {
	return (
		<div className="lusion-canvas-wrap" aria-hidden="true">
			<Canvas
				dpr={[1, 1.75]}
				camera={{ position: [0, 0, 6.6], fov: 42, near: 0.1, far: 40 }}
				gl={{
					antialias: true,
					alpha: true,
					powerPreference: "high-performance",
				}}
			>
				<AcousticScene telemetryRef={telemetryRef} />
			</Canvas>
		</div>
	);
}
