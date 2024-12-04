import {
    OpenContentFavorite,
    OpenContentItem,
    UserRole,
    HelpfulLink
} from '@/common';
import OpenContentCard from '@/Components/cards/OpenContentCard';
import HelpfulLinkCard from '@/Components/cards/HelpfulLinkCard';
import ULIComponent from '@/Components/ULIComponent';
import { useAuth } from '@/useAuth';
import { ArrowTopRightOnSquareIcon } from '@heroicons/react/24/outline';
import { useLoaderData, useNavigate } from 'react-router-dom';
import API from '@/api/api';

export default function OpenContentLevelDashboard() {
    const { user } = useAuth();
    const navigate = useNavigate();
    const { helpfulLinks, topUserContent, topFacilityContent, favorites } =
        useLoaderData() as {
            helpfulLinks: HelpfulLink[];
            topUserContent: OpenContentItem[];
            topFacilityContent: OpenContentItem[];
            favorites: OpenContentFavorite[];
        };

    function navigateToOpenContent() {
        if (user?.role == UserRole.Student) {
            navigate(`/open-content/libraries`);
        } else {
            navigate(`/open-content-management/libraries`);
        }
    }
    async function handleHelpfulLinkClick(id: number): Promise<void> {
        const response = await API.put<{ url: string }, null>(
            `/helpful-links/activity/${id}`,
            null
        );
        if (response.success) {
            if (Array.isArray(response.data)) {
                console.error('unexpected response data');
            } else if (response.data?.url) {
                window.open(response.data.url, '_blank');
            } else {
                console.error('unexpected response data');
            }
        }
    }

    return (
        <div className="flex flex-row h-full">
            {/* main section */}
            <div className="w-full flex flex-col gap-6 px-6 pb-4">
                <h1 className="text-5xl">
                    Hi, {user?.name_first ?? 'Student'}!
                </h1>
                <h2> Pick Up Where You Left Off</h2>
                <div className="grid grid-cols-2 gap-6">
                    <div className="card card-row-padding flex flex-col gap-3">
                        <h2>Your Top Open Content</h2>
                        {topUserContent.map((item: OpenContentItem) => {
                            return (
                                <OpenContentCard
                                    key={item.content_id}
                                    content={item}
                                />
                            );
                        })}
                        {topUserContent.length < 5 && (
                            <div
                                className="card px-4 py-2 flex flex-row gap-2 items-center"
                                onClick={navigateToOpenContent}
                            >
                                <ULIComponent
                                    tooltipClassName="h-12 flex items-center"
                                    iconClassName="w-5 h-5"
                                    icon={ArrowTopRightOnSquareIcon}
                                />
                                <h3>Explore open content offered</h3>
                            </div>
                        )}
                    </div>
                    <div className="card card-row-padding flex flex-col gap-3">
                        <h2>Popular Open Content</h2>
                        {topFacilityContent.map((item: OpenContentItem) => {
                            return (
                                <OpenContentCard
                                    key={item.content_id}
                                    content={item}
                                />
                            );
                        })}
                    </div>
                </div>
                <h2>Resources</h2>
                <div className="card card-row-padding overflow-x-scroll no-scrollbar">
                    {helpfulLinks.map((link: HelpfulLink) => (
                        <a
                            key={link.id}
                            href={link.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="w-[252px] no-underline"
                            onClick={(e) => {
                                e.preventDefault();
                                void handleHelpfulLinkClick(link.id);
                            }}
                        >
                            <div key={link.id} className="w-[252px]">
                                <HelpfulLinkCard
                                    link={link}
                                    role={UserRole.Student}
                                />
                            </div>
                        </a>
                    ))}
                </div>
            </div>
            {/* right sidebar */}
            <div className="min-w-[300px] border-l border-grey-1 flex flex-col gap-6 px-6 py-4">
                <h2>Favorites</h2>
                <div className="space-y-3 w-full">
                    {favorites ? (
                        favorites.map((favorite: OpenContentFavorite) => {
                            return (
                                <OpenContentCard
                                    key={favorite.content_id}
                                    content={favorite}
                                />
                            );
                        })
                    ) : (
                        <div>No Favorites</div>
                    )}
                </div>
            </div>
        </div>
    );
}
