import {
    createContext,
    useContext,
    useEffect,
    useState,
    ReactNode
} from 'react';
import API from '@/api/api';
interface SortOrderContextType {
    globalSortOrder: string;
    setGlobalSortOrder: (sortOrder: string) => void;
}

const SortOrderContext = createContext<SortOrderContextType | undefined>(
    undefined
);

export const SortOrderProvider = ({ children }: { children: ReactNode }) => {
    const [globalSortOrder, setGlobalSortOrder] =
        useState<string>('created_at DESC');

    useEffect(() => {
        async function fetchSortOrder() {
            const response = await API.get<{ sort_order: string }>(
                'helpful-links/sort'
            );
            if (
                response.success &&
                typeof response.data === 'object' &&
                'sort_order' in response.data
            ) {
                setGlobalSortOrder(response.data.sort_order);
            }
        }

        void fetchSortOrder();
    }, []);

    return (
        <SortOrderContext.Provider
            value={{ globalSortOrder, setGlobalSortOrder }}
        >
            {children}
        </SortOrderContext.Provider>
    );
};

export const useSortOrder = () => {
    const context = useContext(SortOrderContext);
    if (!context) {
        throw new Error('useSortOrder must be used within a SortOrderProvider');
    }
    return context;
};
