import { useState, MouseEvent } from 'react';
import VisibleHiddenToggle from './VisibleHiddenToggle';
import { LibraryDto, ServerResponseMany, ToastState, UserRole } from '@/common';
import { StarIcon } from '@heroicons/react/24/solid';
import { StarIcon as StarIconOutline } from '@heroicons/react/24/outline';
import API from '@/api/api';
import { KeyedMutator } from 'swr';
import { useNavigate } from 'react-router-dom';
import { useToast } from '@/Context/ToastCtx';

export default function LibraryCard({
    library,
    mutate,
    role
}: {
    library: LibraryDto;
    mutate: KeyedMutator<ServerResponseMany<LibraryDto>>;
    role: UserRole;
}) {
    const { toaster } = useToast();
    const [visible, setVisible] = useState<boolean>(library.visibility_status);
    const [favorited, setFavorited] = useState<boolean>(
        library.is_favorited ?? false
    );
    const navigate = useNavigate();

    function toggleFavorite(e: MouseEvent) {
        e.preventDefault();
        API.put(`open-content/${library.id}/save`, {
            name: library.name,
            content_url: `/api/proxy/libraries/${library.id}/`,
            open_content_provider_id: library.open_content_provider_id
        })
            .then(() => setFavorited(!favorited))
            .catch((error) => console.error(error));
        library.is_favorited = !favorited;
        void mutate();
    }

    function changeVisibility(visibilityStatus: boolean) {
        if (visibilityStatus !== visible) {
            setVisible(visibilityStatus);
            void handleToggleVisibility();
        }
    }

    const handleToggleVisibility = async () => {
        const response = await API.put<null, object>(
            `libraries/${library.id}`,
            {}
        );
        if (response.success) {
            toaster(response.message, ToastState.success);
            await mutate();
        } else {
            toaster(response.message, ToastState.error);
        }
    };

    const openContentProviderName =
        library?.open_content_provider_name.charAt(0).toUpperCase() +
        library?.open_content_provider_name.slice(1);

    const favoriteIcon = favorited ? (
        <StarIcon
            className="w-5 text-primary-yellow cursor-pointer"
            onClick={toggleFavorite}
        />
    ) : (
        <StarIconOutline
            className="w-5 cursor-pointer"
            onClick={toggleFavorite}
        />
    );

    return (
        <div
            className="card overflow-hidden cursor-pointer"
            onClick={() => navigate(`/viewer/libraries/${library.id}`)}
        >
            <div className="flex p-4 gap-2 border-b-2">
                <figure className="w-[48px] h-[48px] bg-cover">
                    <img
                        src={library.image_url ?? ''}
                        alt={`${library.name} thumbnail`}
                    />
                </figure>
                <h3 className="w-3/4 body my-auto">{library.name}</h3>
                <div onClick={(e) => e.stopPropagation()}>{favoriteIcon}</div>
            </div>
            <div className="p-4 space-y-2">
                <p className="body-small">{openContentProviderName}</p>
                <p className="body-small h-[40px] leading-5 line-clamp-2">
                    {library?.description}
                </p>
                {role === UserRole.Admin && (
                    <VisibleHiddenToggle
                        visible={visible}
                        changeVisibility={changeVisibility}
                    />
                )}
            </div>
        </div>
    );
}
