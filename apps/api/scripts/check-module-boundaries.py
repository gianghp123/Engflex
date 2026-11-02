"""Guard dependency ownership without introducing a DI framework or code layer."""
import json
from pathlib import Path
import re
import sys

ROOT = Path(__file__).resolve().parents[1]
MANIFEST = json.loads((ROOT / "ownership.json").read_text(encoding="utf-8"))
OWNERS = {module: owner for owner, modules in MANIFEST["owners"].items() for module in modules}
REPOSITORIES = {name + "Repository": owner for owner, names in MANIFEST["repositories"].items() for name in names}
errors = []
for path in (ROOT / "internal").rglob("*.go"):
    relative = path.relative_to(ROOT).as_posix()
    if path.name.endswith("_test.go") or "/tests/" in relative or "/mocks/" in relative:
        continue
    source = path.read_text(encoding="utf-8")
    parts = relative.split("/")
    module = parts[2] if relative.startswith("internal/modules/") else None
    if module:
        if module not in OWNERS:
            errors.append(f"{relative}: module has no owner")
            continue
        for target, layer in re.findall(r'"engflex-api/internal/modules/([^/]+)/(services|dtos)[^"\n]*"', source):
            if target != module:
                errors.append(f"{relative}: sibling {target}/{layer} import")
            elif layer == "dtos" and "/services/" in relative and "/dtos/responses" in source:
                errors.append(f"{relative}: response mapping belongs in controllers")
        for repository in set(re.findall(r"repositories\.(\w+Repository)", source)):
            owner = REPOSITORIES.get(repository)
            if owner is None or owner != OWNERS[module]:
                errors.append(f"{relative}: {repository} crosses owner; inject a declared consumer interface")
        if path.name == "module.go" and re.search(r"repositories\.|services\.New|controllers\.New|gorm\.", source):
            errors.append(f"{relative}: module registration constructs dependencies")
    elif not relative.startswith("internal/composition/") and re.search(r'"engflex-api/internal/modules/[^/]+/dtos/', source):
        errors.append(f"{relative}: shared runtime imports a feature wire DTO")
for relative, interfaces in MANIFEST["read_interfaces"].items():
    source = (ROOT / "internal/modules" / relative).read_text(encoding="utf-8")
    for name, expected in interfaces.items():
        match = re.search(r"type " + name + r" interface\s*\{(.*?)\}", source, re.S)
        methods = re.findall(r"^\s*(\w+)\(", match[1], re.M) if match else []
        if set(methods) != set(expected):
            errors.append(f"{relative}: {name} differs from declared scope {expected}")
if errors:
    print("\n".join(errors))
    sys.exit(1)
print(f"Module boundaries passed: {len(OWNERS)} modules, {len(REPOSITORIES)} repository owners")
