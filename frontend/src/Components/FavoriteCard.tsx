import { CombinedFavorite } from '@/common';
import { StarIcon as SolidStar } from '@heroicons/react/24/solid';
import { useNavigate } from 'react-router-dom';

interface FavoriteCardProps {
    favorite: CombinedFavorite;
    onUnfavorite: () => void;
}

const FavoriteCard: React.FC<FavoriteCardProps> = ({
    favorite,
    onUnfavorite
}) => {
    const navigate = useNavigate();

    const handleCardClick = () => {
        if (favorite.is_disabled) {
            return;
        }
        if (favorite.type === 'video') {
            navigate(`/viewer/videos/${favorite.content_id}`);
        } else if (favorite.type === 'library') {
            navigate(`/viewer/libraries/${favorite.content_id}`);
        }
    };

    return (
        <div
            className={`relative rounded-lg p-3 shadow-md transition-all ${
                favorite.is_disabled ? 'bg-grey-2' : 'bg-inner-background'
            } ${
                favorite.is_disabled
                    ? 'cursor-not-allowed'
                    : 'cursor-pointer hover:shadow-lg hover:scale-105'
            }`}
            style={{ width: '220px' }}
            onClick={favorite.is_disabled ? undefined : handleCardClick}
        >
            <div
                className="absolute top-2 right-2 cursor-pointer"
                onClick={(e) => {
                    e.stopPropagation();
                    onUnfavorite();
                }}
            >
                <SolidStar className="w-5 text-primary-yellow" />
            </div>

            <figure className="w-1/2 mx-auto object-contain bg-contain">
                <img
                    src={favorite.thumbnail_url}
                    alt={favorite.name}
                    className="w-full h-28 object-contain rounded-md mb-3"
                />
            </figure>
            <h3 className="text-lg font-bold text-header-text text-center mb-1">
                {favorite.name}
            </h3>
            <p className="text-sm text-body-text text-center">
                {favorite.type === 'video'
                    ? favorite.channel_title
                    : favorite.provider_name}
            </p>
        </div>
    );
};

export default FavoriteCard;
