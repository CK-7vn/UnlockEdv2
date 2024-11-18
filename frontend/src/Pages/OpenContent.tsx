import { OpenContentProviderType, UserRole, Tab } from '@/common';
import { usePathValue } from '@/Context/PathValueCtx';
import { useEffect, useState } from 'react';
import TabView from '@/Components/TabView';
import { useNavigate, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '@/useAuth';
import { mutate } from 'swr';

export default function OpenContent() {
    const { setPathVal } = usePathValue();
    const navigate = useNavigate();
    const { user } = useAuth();
    const route = useLocation();
    const tab = route.pathname.split('/')[2] ?? 'libraries';
    const [activeTab, setActiveTab] = useState<Tab>(
        tab.toLowerCase() === 'libraries'
            ? {
                  name: 'Kiwix',
                  value: 'Libraries'
              }
            : { name: 'Videos', value: 'Videos' }
    );
    useEffect(() => {
        void setPathVal([
            { path_id: ':kind', value: activeTab.value as string }
        ]);
    }, [activeTab]);
    const tabs = [
        { name: OpenContentProviderType.KIWIX, value: 'Libraries' },
        { name: OpenContentProviderType.VIDEOS, value: 'Videos' },
        { name: 'Favorites', value: 'Favorites' }
    ];

    const handlePageChange = (tab: Tab) => {
        setActiveTab(tab);
        navigate(`/open-content/${tab.value}`);

        if (tab.value === 'Libraries' || tab.value === 'Videos') {
            void mutate('/api/libraries');
            void mutate('open-content/favorites');
            void setPathVal([{ path_id: ':kind', value: tab.value as string }]);
        }
    };

    return (
        <div className="px-8 pb-4">
            <div className="flex flex-row justify-between">
                <h1>Open Content</h1>
                {user?.role === UserRole.Admin && (
                    <button
                        className="button border border-primary bg-transparent text-body-text"
                        onClick={() =>
                            navigate(
                                `/open-content-management/${activeTab.value}`
                            )
                        }
                    >
                        Return to Admin View
                    </button>
                )}
            </div>
            <TabView
                tabs={tabs}
                activeTab={activeTab}
                setActiveTab={(tab) => {
                    void handlePageChange(tab);
                }}
            />
            <div className="flex flex-row gap-4 pt-8 pb-8">
                <Outlet />
            </div>
        </div>
    );
}
