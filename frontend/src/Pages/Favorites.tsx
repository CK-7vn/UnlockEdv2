import { useEffect, useState } from 'react';
import { useAuth } from '@/useAuth';
import API from '@/api/api';
import FavoriteCard from '@/Components/FavoriteCard';
import { CombinedFavorite } from '@/common';
import Loading from '@/Components/Loading';
import Error from '@/Pages/Error';
import { mutate } from 'swr';

export default function FavoritesPage() {
    const { user } = useAuth();
    const [favorites, setFavorites] = useState<CombinedFavorite[]>([]);
    const [isLoading, setIsLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchFavorites = async (): Promise<void> => {
            try {
                setIsLoading(true);
                const response = await API.get('open-content/favorites');
                interface ApiResponse {
                    data: {
                        id: number;
                        content_id: number;
                        name: string;
                        type: 'library' | 'video';
                        thumbnail_url: string;
                        description: string;
                        is_disabled: boolean;
                        open_content_provider_id: number;
                        channel_title?: string;
                        provider_name?: string;
                    }[];
                }
                const typedResponse = response as ApiResponse;
                const mappedFavorites: CombinedFavorite[] =
                    typedResponse.data.map((item) => ({
                        id: item.id,
                        content_id: item.content_id,
                        name: item.name,
                        type: item.type,
                        thumbnail_url: item.thumbnail_url,
                        description: item.description,
                        is_disabled: item.is_disabled,
                        open_content_provider_id: item.open_content_provider_id,
                        channel_title: item.channel_title ?? '',
                        provider_name: item.provider_name ?? ''
                    }));
                setFavorites(mappedFavorites);
            } catch (err) {
                setError('Error fetching favorites.');
                console.error(err);
            } finally {
                setIsLoading(false);
            }
        };
        if (user) {
            void fetchFavorites();
        }
    }, [user]);

    const handleUnfavorite = async (
        contentId: number,
        type: 'video' | 'library'
    ) => {
        try {
            setIsLoading(true);
            let endpoint = '';

            if (type === 'video') {
                endpoint = `videos/${contentId}/favorite`;
            } else {
                endpoint = `open-content/${contentId}/save`;
            }

            const favorite = favorites.find(
                (fav) => fav.content_id === contentId
            );

            if (!favorite) {
                setError('Favorite not found.');
                return;
            }

            const { name, open_content_provider_id } = favorite;
            if (!name || !open_content_provider_id) {
                setError('Favorite data is incomplete.');
                return;
            }

            const payload = {
                name,
                content_url: `/api/proxy/libraries/${favorite.content_id}/`,
                open_content_provider_id
            };

            const response = await API.put(endpoint, payload);

            if (response.success) {
                setFavorites((prevFavorites) =>
                    prevFavorites.filter((fav) => fav.content_id !== contentId)
                );
                if (type === 'library') {
                    await mutate('/api/libraries');
                }
            } else {
                setError('Failed to unfavorite item.');
            }
        } catch (err) {
            setError('Error unfavoriting item.');
            console.error(err);
        } finally {
            setIsLoading(false);
        }
    };

    if (isLoading) {
        return <Loading />;
    }

    if (error) {
        return <Error />;
    }

    return (
        <div className="px-8 pb-4">
            <h1>My Favorites</h1>
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6 mt-6">
                {favorites.length > 0 ? (
                    favorites.map((favorite) => (
                        <FavoriteCard
                            key={favorite.content_id}
                            favorite={favorite}
                            onUnfavorite={() =>
                                void handleUnfavorite(
                                    favorite.content_id,
                                    favorite.type
                                )
                            }
                        />
                    ))
                ) : (
                    <p>You have no favorites yet.</p>
                )}
            </div>
        </div>
    );
}
