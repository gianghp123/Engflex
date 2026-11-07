import { Link } from "@tanstack/react-router";
import { cn } from "cn";
import {
	BookMarked,
	BookmarkCheck,
	BookOpen,
	Calendar,
	Check,
	Clapperboard,
	ExternalLink,
	Pencil,
	RotateCw,
	StickyNote,
	Trash2,
	X,
} from "lucide-react";
import { type MouseEvent, useState } from "react";
import { toast } from "sonner";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import { Textarea } from "#/components/ui/textarea";
import { m } from "#/paraglide/messages";
import { notesActions } from "../store";
import type { NoteItem, NoteKind } from "../types";

const KIND_CONFIG: Record<
	NoteKind,
	{
		label: () => string;
		icon: typeof BookmarkCheck;
		badgeClass: string;
		accentBorder: string;
		glowColor: string;
	}
> = {
	vocab: {
		label: () => m["notes.tabs.vocab"](),
		icon: BookmarkCheck,
		badgeClass:
			"bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/25",
		accentBorder: "group-hover:border-emerald-500/40",
		glowColor: "rgba(16, 185, 129, 0.12)",
	},
	lesson: {
		label: () => m["notes.tabs.lesson"](),
		icon: BookOpen,
		badgeClass:
			"bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/25",
		accentBorder: "group-hover:border-blue-500/40",
		glowColor: "rgba(59, 130, 246, 0.12)",
	},
	video: {
		label: () => m["notes.tabs.video"](),
		icon: Clapperboard,
		badgeClass:
			"bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/25",
		accentBorder: "group-hover:border-rose-500/40",
		glowColor: "rgba(244, 63, 94, 0.12)",
	},
	story: {
		label: () => m["notes.tabs.story"](),
		icon: BookMarked,
		badgeClass:
			"bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/25",
		accentBorder: "group-hover:border-amber-500/40",
		glowColor: "rgba(245, 158, 11, 0.12)",
	},
};

export function NoteCard3D({ item }: { item: NoteItem }) {
	const [isFlipped, setIsFlipped] = useState(false);
	const [isEditing, setIsEditing] = useState(false);
	const [draftNote, setDraftNote] = useState(item.personalNote ?? "");

	const [tilt, setTilt] = useState({
		rotateX: 0,
		rotateY: 0,
		glareX: 50,
		glareY: 50,
		opacity: 0,
	});

	const config = KIND_CONFIG[item.kind];
	const Icon = config.icon;

	const handleMouseMove = (e: MouseEvent<HTMLDivElement>) => {
		if (isFlipped) return; // Giữ phẳng khi đang lật mặt sau để người dùng đọc/sửa note thoải mái
		const rect = e.currentTarget.getBoundingClientRect();
		const x = e.clientX - rect.left;
		const y = e.clientY - rect.top;
		const centerX = rect.width / 2;
		const centerY = rect.height / 2;

		const rotateX = -((y - centerY) / centerY) * 8; // tối đa 8 độ
		const rotateY = ((x - centerX) / centerX) * 8;

		setTilt({
			rotateX,
			rotateY,
			glareX: (x / rect.width) * 100,
			glareY: (y / rect.height) * 100,
			opacity: 1,
		});
	};

	const handleMouseLeave = () => {
		setTilt({
			rotateX: 0,
			rotateY: 0,
			glareX: 50,
			glareY: 50,
			opacity: 0,
		});
	};

	const handleSaveNote = () => {
		notesActions.updatePersonalNote(item.id, draftNote);
		setIsEditing(false);
		toast.success(m["notes.toast.saved"]());
	};

	const handleCancelEdit = () => {
		setDraftNote(item.personalNote ?? "");
		setIsEditing(false);
	};

	const handleRemove = () => {
		notesActions.removeItem(item.id);
		toast.success(m["notes.toast.removed"]());
	};

	return (
		// biome-ignore lint/a11y/noStaticElementInteractions: visual-only 3D mouse tracking tilt
		<div
			className="group relative h-[380px] w-full [perspective:1200px]"
			onMouseMove={handleMouseMove}
			onMouseLeave={handleMouseLeave}
		>
			<div
				className={cn(
					"relative h-full w-full rounded-2xl transition-all duration-500 [transform-style:preserve-3d]",
				)}
				style={{
					transform: isFlipped
						? "rotateY(180deg)"
						: `rotateX(${tilt.rotateX}deg) rotateY(${tilt.rotateY}deg)`,
					boxShadow: isFlipped
						? "0 12px 30px -10px rgba(0, 0, 0, 0.2)"
						: `0 14px 32px -8px ${config.glowColor}, 0 6px 16px -4px rgba(0, 0, 0, 0.08)`,
				}}
			>
				{/* ---------------- MẶT TRƯỚC (FRONT FACE) ---------------- */}
				<div
					className={cn(
						"surface-card absolute inset-0 flex flex-col justify-between overflow-hidden rounded-2xl border p-5 transition-colors [backface-visibility:hidden]",
						config.accentBorder,
					)}
				>
					{/* Hiệu ứng specular glare (bóng kính 3D) */}
					<div
						className="pointer-events-none absolute inset-0 rounded-2xl opacity-0 transition-opacity duration-300 group-hover:opacity-100"
						style={{
							background: `radial-gradient(circle at ${tilt.glareX}% ${tilt.glareY}%, rgba(255, 255, 255, 0.18) 0%, transparent 65%)`,
							opacity: tilt.opacity,
						}}
					/>

					{/* Top bar: Badge loại & Quick actions */}
					<div>
						<div className="flex items-center justify-between gap-2">
							<Badge
								variant="outline"
								className={cn(
									"flex items-center gap-1.5 px-2.5 py-1 font-semibold",
									config.badgeClass,
								)}
							>
								<Icon className="size-3.5" />
								<span>{config.label()}</span>
							</Badge>

							<div className="flex items-center gap-1">
								{item.extraInfo?.level && (
									<Badge
										variant="secondary"
										className="font-mono text-xs font-bold"
									>
										{item.extraInfo.level}
									</Badge>
								)}
								{item.extraInfo?.duration && (
									<Badge variant="secondary" className="text-xs">
										{item.extraInfo.duration}
									</Badge>
								)}
								{item.extraInfo?.badge && !item.extraInfo.level && (
									<Badge variant="outline" className="text-xs">
										{item.extraInfo.badge}
									</Badge>
								)}

								<Button
									variant="ghost"
									size="icon"
									className="size-8 text-muted-foreground hover:text-destructive"
									onClick={handleRemove}
									aria-label={m["notes.actions.unsave"]()}
								>
									<Trash2 className="size-4" />
								</Button>
							</div>
						</div>

						{/* Tiêu đề & phụ đề */}
						<div className="mt-3">
							<h3 className="line-clamp-2 text-lg font-bold text-foreground">
								{item.title}
							</h3>
							{item.subtitle && (
								<p className="mt-0.5 text-xs font-medium text-muted-foreground">
									{item.subtitle}
								</p>
							)}
						</div>

						{/* Tóm tắt / Nghĩa */}
						{item.summary && (
							<p className="mt-2.5 line-clamp-2 text-xs leading-relaxed text-muted-foreground">
								{item.summary}
							</p>
						)}

						{/* Khối xem nhanh ghi chú (Quick Note Preview) */}
						<button
							type="button"
							onClick={() => setIsFlipped(true)}
							className="mt-3 w-full text-left cursor-pointer rounded-xl border border-primary/20 bg-primary/5 p-2.5 transition-colors hover:bg-primary/10"
						>
							<div className="flex items-center justify-between text-xs font-semibold text-primary">
								<span className="flex items-center gap-1.5">
									<StickyNote className="size-3.5" />
									{m["notes.card.personalNoteLabel"]()}
								</span>
								<span className="flex items-center gap-1 text-[11px] font-normal text-muted-foreground">
									<RotateCw className="size-3" />
									{m["notes.actions.flipNote"]()}
								</span>
							</div>
							<p className="mt-1 line-clamp-2 text-xs text-foreground/90 italic">
								{item.personalNote || m["notes.card.emptyNote"]()}
							</p>
						</button>
					</div>

					{/* Bottom section: Tags & Nav buttons */}
					<div className="mt-3 pt-3 border-t">
						<div className="flex flex-wrap gap-1 mb-3">
							{item.tags.slice(0, 3).map((tag) => (
								<span
									key={tag}
									className="rounded-md bg-muted px-2 py-0.5 text-[11px] font-medium text-muted-foreground"
								>
									#{tag}
								</span>
							))}
						</div>

						<div className="flex items-center gap-2">
							<Button
								variant="outline"
								size="sm"
								className="flex-1 gap-1.5 text-xs font-medium"
								onClick={() => setIsFlipped(true)}
							>
								<RotateCw className="size-3.5" />
								{m["notes.actions.flipNote"]()}
							</Button>

							<Button asChild size="sm" className="gap-1.5 text-xs font-medium">
								<Link to={item.linkTo}>
									{m["notes.actions.openContent"]()}
									<ExternalLink className="size-3.5" />
								</Link>
							</Button>
						</div>
					</div>
				</div>

				{/* ---------------- MẶT SAU (BACK FACE — 3D FLIPPED) ---------------- */}
				<div
					className={cn(
						"surface-card absolute inset-0 flex flex-col justify-between overflow-hidden rounded-2xl border p-5 [backface-visibility:hidden] [transform:rotateY(180deg)]",
						config.accentBorder,
					)}
				>
					{/* Header mặt sau */}
					<div>
						<div className="flex items-center justify-between border-b pb-2.5">
							<div className="flex items-center gap-2">
								<StickyNote className="size-4 text-primary" />
								<span className="text-sm font-bold text-foreground">
									{m["notes.card.personalNoteLabel"]()}
								</span>
							</div>

							<Button
								variant="ghost"
								size="sm"
								className="h-7 gap-1 px-2 text-xs text-muted-foreground hover:text-foreground"
								onClick={() => {
									setIsFlipped(false);
									setIsEditing(false);
								}}
							>
								<RotateCw className="size-3" />
								{m["notes.actions.flipBack"]()}
							</Button>
						</div>

						<div className="mt-2.5">
							<p className="line-clamp-1 text-xs font-semibold text-foreground/80">
								{item.title}
							</p>
							<div className="mt-1 flex items-center gap-1.5 text-[11px] text-muted-foreground">
								<Calendar className="size-3" />
								<span>
									{m["notes.card.savedDatePrefix"]()} {item.createdAt}
								</span>
							</div>
						</div>

						{/* Vùng chỉnh sửa / Xem ghi chú */}
						<div className="mt-3">
							{isEditing ? (
								<div className="space-y-2">
									<Textarea
										value={draftNote}
										onChange={(e) => setDraftNote(e.target.value)}
										placeholder={m["notes.dialog.notePlaceholder"]()}
										className="min-h-[140px] resize-none text-xs"
									/>
									<div className="flex justify-end gap-1.5">
										<Button
											variant="ghost"
											size="sm"
											className="h-7 text-xs"
											onClick={handleCancelEdit}
										>
											<X className="mr-1 size-3" />
											{m["notes.actions.cancel"]()}
										</Button>
										<Button
											size="sm"
											className="h-7 text-xs"
											onClick={handleSaveNote}
										>
											<Check className="mr-1 size-3" />
											{m["notes.actions.save"]()}
										</Button>
									</div>
								</div>
							) : (
								<div className="relative min-h-[140px] rounded-xl border bg-muted/40 p-3">
									<p className="whitespace-pre-wrap text-xs leading-relaxed text-foreground">
										{item.personalNote || (
											<span className="text-muted-foreground italic">
												{m["notes.card.emptyNote"]()}
											</span>
										)}
									</p>
									<Button
										variant="ghost"
										size="icon"
										className="absolute right-2 top-2 size-7 text-muted-foreground hover:text-foreground"
										onClick={() => setIsEditing(true)}
										aria-label={m["notes.actions.editNote"]()}
									>
										<Pencil className="size-3.5" />
									</Button>
								</div>
							)}
						</div>
					</div>

					{/* Footer mặt sau */}
					<div className="mt-3 flex items-center justify-between border-t pt-3">
						<Button
							variant="outline"
							size="sm"
							className="h-8 gap-1.5 text-xs text-destructive hover:bg-destructive/10"
							onClick={handleRemove}
						>
							<Trash2 className="size-3.5" />
							{m["notes.actions.unsave"]()}
						</Button>

						<Button asChild size="sm" className="h-8 gap-1.5 text-xs">
							<Link to={item.linkTo}>
								{m["notes.actions.openContent"]()}
								<ExternalLink className="size-3.5" />
							</Link>
						</Button>
					</div>
				</div>
			</div>
		</div>
	);
}
