import {
    LibraryAdminVisibility,
    Library,
    ServerResponseMany,
    UserRole,
    OpenContentCategory
} from '@/common';
import DropdownControl from '@/Components/inputs/DropdownControl';
import SearchBar from '@/Components/inputs/SearchBar';
import LibraryCard from '@/Components/LibraryCard';
import { isAdministrator, useAuth } from '@/useAuth';
import { useEffect, useState } from 'react';
import useSWR from 'swr';
import Pagination from './Pagination';
import { AxiosError } from 'axios';
import { useLoaderData, useLocation } from 'react-router-dom';

export default function LibaryLayout({
    studentView
}: {
    studentView?: boolean;
}) {
    const { user } = useAuth();
    if (!user) {
        return null;
    }
    const { categories } = useLoaderData() as {
        categories: OpenContentCategory[];
    };
    console.log(categories);
    const categoriesFilter = transformArrayToObject(categories);
    function transformArrayToObject(
        categories: OpenContentCategory[]
    ): Record<string, string> {
        return categories.reduce(
            (acc, category, currentIndex) => {
                if (currentIndex == 0) {
                    acc['All Libraries'] = '';
                }
                acc[category.name] = category.id.toString();
                return acc;
            },
            {} as Record<string, string>
        );
    }
    const [searchTerm, setSearchTerm] = useState<string>('');
    const [filterLibraries, setFilterLibraries] = useState<string>(
        categoriesFilter['All Libraries']
    );
    const [filterVisibilityAdmin, setFilterVisibilityAdmin] = useState<string>(
        LibraryAdminVisibility['All Libraries']
    );
    let role = user.role;
    if (studentView) {
        role = UserRole.Student;
    }
    const [perPage, setPerPage] = useState(20);
    const [pageQuery, setPageQuery] = useState<number>(1);
    const route = useLocation();
    const adminWithStudentView = (): boolean => {
        return !route.pathname.includes('management') && isAdministrator(user);
    };
    const {
        data: libraries,
        mutate: mutateLibraries,
        error: librariesError,
        isLoading: librariesLoading
    } = useSWR<ServerResponseMany<Library>, AxiosError>(
        `/api/libraries?page=${pageQuery}&per_page=${perPage}&visibility=${isAdministrator(user) && !adminWithStudentView() ? filterVisibilityAdmin : 'visible'}&search=${searchTerm}&category=${filterLibraries}`
    );
    const librariesMeta = libraries?.meta ?? {
        total: 0,
        per_page: 20,
        page: 1,
        current_page: 1,
        last_page: 1
    };

    const handleSetPerPage = (perPage: number) => {
        setPerPage(perPage);
        setPageQuery(1);
        void mutateLibraries();
    };

    useEffect(() => {
        setPageQuery(1);
    }, [filterVisibilityAdmin, filterLibraries, searchTerm]);

    function updateLibrary() {
        void mutateLibraries();
    }

    return (
        <>
            <div className="flex flex-row gap-4">
                <SearchBar
                    searchTerm={searchTerm}
                    changeCallback={setSearchTerm}
                />
                {isAdministrator(user) && !adminWithStudentView() && (
                    <DropdownControl
                        enumType={LibraryAdminVisibility}
                        setState={setFilterVisibilityAdmin}
                    />
                )}
                <DropdownControl
                    enumType={categoriesFilter}
                    setState={setFilterLibraries}
                />
            </div>
            <div className="grid grid-cols-4 gap-6">
                {libraries?.data.map((library) => (
                    <LibraryCard
                        key={library.id}
                        library={library}
                        mutate={updateLibrary}
                        role={adminWithStudentView() ? UserRole.Student : role}
                    />
                ))}
            </div>
            {!librariesLoading && !librariesError && librariesMeta && (
                <div className="flex justify-center">
                    <Pagination
                        meta={librariesMeta}
                        setPage={setPageQuery}
                        setPerPage={handleSetPerPage}
                    />
                </div>
            )}
        </>
    );
}
