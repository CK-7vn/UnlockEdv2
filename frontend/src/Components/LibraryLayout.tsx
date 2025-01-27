import {
    LibraryAdminVisibility,
    Library,
    ServerResponseMany,
    UserRole,
    Option
} from '@/common';
import DropdownControl from '@/Components/inputs/DropdownControl';
import MultiSelectDropdown from './MultiSelectDropdownControl';
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
        categories: Option[];
    };
    const [selectedCategories, setSelectedCategories] = useState<number[]>([]);
    const [searchTerm, setSearchTerm] = useState<string>('');
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
        `/api/libraries?page=${pageQuery}&per_page=${perPage}&visibility=${isAdministrator(user) && !adminWithStudentView() ? filterVisibilityAdmin : 'visible'}&search=${searchTerm}&category=${selectedCategories.includes(0) ? '' : selectedCategories.join(',')}`
    );

    const uniqueLibraries = libraries?.data
        ? Array.from(
              new Map(libraries.data.map((lib) => [lib.id, lib])).values()
          )
        : [];
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
    }, [filterVisibilityAdmin, searchTerm, selectedCategories]);

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
                <MultiSelectDropdown
                    label="Categories"
                    options={categories}
                    selectedOptions={selectedCategories}
                    onSelectionChange={setSelectedCategories}
                    onBlurSearch={() => {
                        void mutateLibraries;
                    }}
                />
            </div>
            <div className="grid grid-cols-4 gap-6">
                {uniqueLibraries?.map((library) => (
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
