import { useSortOrder } from '@/Context/SortOrderCtx';
export default function SortByPills({
    label,
    updateSort
}: {
    label: { name: string; value: string };
    updateSort: (value: string) => void;
}) {
    const { globalSortOrder } = useSortOrder();
    const isSelected = globalSortOrder === label.value;
    return (
        <div
            className={`${isSelected ? `bg-teal-1 border-2 border-black shadow-md` : `bg-grey-1`} px-3 py-1 rounded-2xl cursor-pointer body`}
            onClick={() => void updateSort(label.value)}
        >
            {label.name}
        </div>
    );
}
