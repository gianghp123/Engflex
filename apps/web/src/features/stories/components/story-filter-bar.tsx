import { cn } from "cn";
import { Bookmark, BookOpen, Check, ChevronDown, Search } from "lucide-react";
import { useMemo, useState } from "react";
import { Button } from "#/components/ui/button";
import { Input } from "#/components/ui/input";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "#/components/ui/popover";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#/components/ui/select";
import { m } from "#/paraglide/messages";

const LEVELS = ["all", "A1", "A2", "B1", "B2", "C1", "C2"] as const;

function CollectionSearchableSelect({
	categories,
	selectedCategory,
	onCategoryChange,
}: {
	categories: string[];
	selectedCategory: string;
	onCategoryChange: (category: string) => void;
}) {
	const [open, setOpen] = useState(false);
	const [search, setSearch] = useState("");

	const filtered = useMemo(() => {
		if (!search.trim()) return categories;
		const q = search.toLowerCase();
		return categories.filter((c) => c.toLowerCase().includes(q));
	}, [categories, search]);

	const selectedLabel =
		selectedCategory === "all"
			? m["stories.filter.allCollections"]()
			: selectedCategory;

	return (
		<Popover open={open} onOpenChange={setOpen}>
			<PopoverTrigger asChild>
				<Button
					type="button"
					variant="outline"
					className="h-9 w-44 justify-between rounded-xl border-border bg-card px-3 text-xs font-semibold hover:bg-muted"
				>
					<span className="truncate">{selectedLabel}</span>
					<ChevronDown className="size-4 shrink-0 opacity-50" />
				</Button>
			</PopoverTrigger>
			<PopoverContent align="start" className="w-56 p-2">
				<div className="relative mb-2">
					<Search className="absolute top-1/2 left-2.5 size-3.5 -translate-y-1/2 text-muted-foreground" />
					<Input
						value={search}
						onChange={(e) => setSearch(e.target.value)}
						placeholder={m["stories.filter.searchCollectionPlaceholder"]()}
						className="h-8 rounded-lg pl-8 text-xs"
					/>
				</div>
				<div className="flex max-h-48 flex-col gap-0.5 overflow-y-auto">
					<button
						type="button"
						onClick={() => {
							onCategoryChange("all");
							setOpen(false);
							setSearch("");
						}}
						className={cn(
							"flex w-full items-center justify-between rounded-lg px-2.5 py-1.5 text-left text-xs font-medium transition-colors",
							selectedCategory === "all"
								? "bg-accent font-bold text-accent-foreground"
								: "text-foreground hover:bg-muted",
						)}
					>
						<span>{m["stories.filter.allCollections"]()}</span>
						{selectedCategory === "all" ? (
							<Check className="size-3.5 text-primary" />
						) : null}
					</button>

					{filtered.map((cat) => {
						const isSelected = selectedCategory === cat;
						return (
							<button
								key={cat}
								type="button"
								onClick={() => {
									onCategoryChange(cat);
									setOpen(false);
									setSearch("");
								}}
								className={cn(
									"flex w-full items-center justify-between rounded-lg px-2.5 py-1.5 text-left text-xs font-medium transition-colors",
									isSelected
										? "bg-accent font-bold text-accent-foreground"
										: "text-foreground hover:bg-muted",
								)}
							>
								<span className="truncate">{cat}</span>
								{isSelected ? (
									<Check className="size-3.5 text-primary" />
								) : null}
							</button>
						);
					})}
				</div>
			</PopoverContent>
		</Popover>
	);
}

export function StoryFilterBar({
	tab,
	onTabChange,
	level,
	onLevelChange,
	categories,
	selectedCategory,
	onCategoryChange,
	search,
	onSearchChange,
}: {
	tab: "all" | "saved";
	onTabChange: (tab: "all" | "saved") => void;
	level: string;
	onLevelChange: (level: string) => void;
	categories: string[];
	selectedCategory: string;
	onCategoryChange: (category: string) => void;
	search: string;
	onSearchChange: (search: string) => void;
}) {
	const [draft, setDraft] = useState(search);
	const [seen, setSeen] = useState(search);
	if (seen !== search) {
		setSeen(search);
		setDraft(search);
	}

	function handleSearchSubmit(e: React.FormEvent) {
		e.preventDefault();
		onSearchChange(draft.trim());
	}

	return (
		<div className="flex flex-col gap-4 border-b border-border pb-5">
			<div className="flex items-center gap-6 border-b border-border/60">
				<button
					type="button"
					onClick={() => onTabChange("all")}
					className={cn(
						"flex items-center gap-2 border-b-2 px-1 pb-3 text-sm font-bold transition-colors",
						tab === "all"
							? "border-primary text-foreground"
							: "border-transparent text-muted-foreground hover:text-foreground",
					)}
				>
					<BookOpen className="size-4" />
					<span>{m["stories.tabs.all"]()}</span>
				</button>
				<button
					type="button"
					onClick={() => onTabChange("saved")}
					className={cn(
						"flex items-center gap-2 border-b-2 px-1 pb-3 text-sm font-bold transition-colors",
						tab === "saved"
							? "border-primary text-foreground"
							: "border-transparent text-muted-foreground hover:text-foreground",
					)}
				>
					<Bookmark className="size-4" />
					<span>{m["stories.tabs.saved"]()}</span>
				</button>
			</div>

			<div className="flex flex-wrap items-center justify-between gap-4 pt-1">
				<div className="flex flex-wrap items-center gap-3">
					<Select value={level} onValueChange={onLevelChange}>
						<SelectTrigger className="h-9 w-36 rounded-xl border-border bg-card px-3 text-xs font-semibold">
							<SelectValue placeholder={m["stories.filter.allLevels"]()} />
						</SelectTrigger>
						<SelectContent>
							{LEVELS.map((lvl) => (
								<SelectItem key={lvl} value={lvl}>
									{lvl === "all" ? m["stories.filter.allLevels"]() : lvl}
								</SelectItem>
							))}
						</SelectContent>
					</Select>

					<CollectionSearchableSelect
						categories={categories}
						selectedCategory={selectedCategory}
						onCategoryChange={onCategoryChange}
					/>
				</div>

				<form
					onSubmit={handleSearchSubmit}
					className="relative w-full max-w-xs"
				>
					<Search className="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
					<Input
						value={draft}
						onChange={(e) => setDraft(e.target.value)}
						onBlur={() => {
							if (draft.trim() !== search) {
								onSearchChange(draft.trim());
							}
						}}
						placeholder={m["stories.filter.searchPlaceholder"]()}
						className="h-9 rounded-xl pl-9 text-xs"
					/>
				</form>
			</div>
		</div>
	);
}
