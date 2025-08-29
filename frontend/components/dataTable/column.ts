import { h } from "vue";
import type { ColumnDef } from "@tanstack/vue-table";

export function generateColumnsFromKeys<T extends { id: string }>(
  keys: (keyof T)[],
  handlers?: {
    onEdit?: (row: T) => void;
    onDelete?: (row: T) => void;
  },
): ColumnDef<T>[] {
  const columns: ColumnDef<T>[] = keys.map((key) => ({
    accessorKey: key,
    header: () => key.toString(),
    cell: ({ row }: any) => row.getValue(key)?.toString() ?? "",
  }));

  // Actions column
  columns.push({
    id: "actions",
    header: "Actions",
    cell: ({ row }: any) =>
      h("div", { class: "flex gap-2" }, [
        h(
          "button",
          {
            class: "px-2 py-1 bg-blue-500 text-white rounded",
            onClick: () => handlers?.onEdit?.(row.original),
          },
          "Edit",
        ),
        h(
          "button",
          {
            class: "px-2 py-1 bg-red-500 text-white rounded",
            onClick: () => handlers?.onDelete?.(row.original),
          },
          "Delete",
        ),
      ]),
  });

  return columns;
}
