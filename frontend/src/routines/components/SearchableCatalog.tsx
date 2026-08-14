import { useMemo, useState, type ReactNode } from "react";
import { Search } from "lucide-react";

interface CatalogItem {
  id: number;
  name: string;
  description?: string;
}

interface SearchableCatalogProps<T extends CatalogItem> {
  label: string;
  items: T[];
  children: (item: T) => ReactNode;
}

export function SearchableCatalog<T extends CatalogItem>({
  label,
  items,
  children,
}: SearchableCatalogProps<T>) {
  const [query, setQuery] = useState("");
  const filtered = useMemo(
    () =>
      items.filter((item) =>
        item.name.toLocaleLowerCase().includes(query.toLocaleLowerCase()),
      ),
    [items, query],
  );
  return (
    <div>
      <label className="grid gap-2">
        <span className="label-text">{label}</span>
        <span className="input input-bordered flex w-full items-center gap-2">
          <Search aria-hidden="true" size={18} />
          <input
            type="search"
            className="grow"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Buscar por nombre"
          />
        </span>
      </label>
      <p className="mt-2 text-sm text-base-content/70" role="status">
        {filtered.length} resultados
      </p>
      {filtered.length === 0 ? (
        <p className="mt-4">No hay coincidencias.</p>
      ) : (
        <div className="mt-3 grid gap-2">
          {filtered.map((item) => (
            <div key={item.id}>{children(item)}</div>
          ))}
        </div>
      )}
    </div>
  );
}
