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
            } hover:shadow-lg hover:scale-105 cursor-pointer`}
            style={{ width: '220px' }}
            onClick={handleCardClick}
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
            <img
                src={favorite.thumbnail_url}
                alt={favorite.name}
                className="w-full h-28 object-cover rounded-md mb-3"
            />
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
