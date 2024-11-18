import { useEffect, useState, MouseEvent, useRef } from 'react';
import { useParams } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import Error from '@/Pages/Error';
import API from '@/api/api';
import { Library, ServerResponseOne, ToastState } from '@/common';
import { usePathValue } from '@/Context/PathValueCtx';
import {
    StarIcon as StarIconOutline,
    StarIcon
} from '@heroicons/react/24/solid';
import { useToast } from '@/Context/ToastCtx';
import Modal from '@/Components/Modal';
import { TextInput } from '@/Components/inputs/TextInput';
import { ModalType } from '@/common';
import { SubmitButton } from '@/Components/inputs';
import { CloseX } from '@/Components/inputs/CloseX';

export default function LibraryViewer() {
    const { id: libraryId } = useParams();
    const [src, setSrc] = useState<string>('');
    const [error, setError] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState<boolean>(true);
    const { setPathVal } = usePathValue();
    const { toaster } = useToast();
    const [favorited, setFavorited] = useState<boolean>(false);
    const [providerID, setProviderID] = useState<number>();
    const favoriteModalRef = useRef<HTMLDialogElement>(null);
    const {
        register,
        handleSubmit,
        formState: { errors },
        reset
    } = useForm<{ favoriteName: string }>();

    useEffect(() => {
        const fetchLibraryData = async () => {
            setIsLoading(true);
            try {
                const resp = (await API.get(
                    `libraries/${libraryId}`
                )) as ServerResponseOne<Library>;
                if (resp.success) {
                    setPathVal([
                        { path_id: ':library_name', value: resp.data.name }
                    ]);
                    setProviderID(resp.data.open_content_provider_id);
                    setFavorited(resp.data.is_favorited ?? false);
                }
                const response = await fetch(
                    `/api/proxy/libraries/${libraryId}/`
                );
                if (response.ok) {
                    setSrc(response.url);
                } else if (response.status === 404) {
                    setError('Library not found');
                } else {
                    setError('Error loading library');
                }
            } catch {
                setError('Error loading library');
            } finally {
                setIsLoading(false);
            }
        };
        void fetchLibraryData();
        return () => {
            sessionStorage.removeItem('tag');
        };
    }, [libraryId]);
    // useRef avoids rerendering possiblities which may cause issues with websocket connection, so that is why i'm using useRef here to hold state in case that happens
    const socketRef = useRef<WebSocket | null>(null);
    const [isConnected, setIsConnected] = useState(false);
    useEffect(() => {
        //websocket effect
        const protocol =
            window.location.protocol === 'https:' ? 'wss://' : 'ws://';
        const host = window.location.hostname;
        const websockeUrl = `${protocol}${host}`;
        const socket = new WebSocket(websockeUrl + '/api/ws/listen');
        // Handle incoming messages
        socketRef.current = socket;
        socket.onopen = () => {
            setIsConnected(true);
        };
        socket.onclose = () => {
            setIsConnected(false);
        };
        socket.onmessage = (event) => {
            try {
                if (event.data == 'true') {
                    setFavorited(true);
                } else if (event.data == 'false') {
                    setFavorited(false);
                }
            } catch (error) {
                console.error('Error parsing WebSocket message:', error);
            }
        };
        //send message to let server know the user?
        return () => {
            if (socketRef.current && isConnected) {
                socketRef.current.close();
            }
        };
    }, []);

    const toggleFavorite = (e: MouseEvent) => {
        e.preventDefault();

        if (favorited) {
            void handleUnfavorite();
        } else {
            favoriteModalRef.current?.showModal();
        }
    };

    const handleUnfavorite = async () => {
        try {
            const response = await API.put(`open-content/${libraryId}/save`, {
                name: ''
            });
            if (response.success) {
                setFavorited(false);
                toaster('Library removed from favorites', ToastState.success);
            } else {
                toaster(response.message, ToastState.error);
            }
        } catch (error: unknown) {
            toaster('Error updating favorite status', ToastState.error);
            console.error(error);
        }
    };

    const handleFavoriteSubmit = async (data: { favoriteName: string }) => {
        try {
            const response = await API.put(`open-content/${libraryId}/save`, {
                name: data.favoriteName,
                open_content_provider_id: providerID
            });
            if (response.success) {
                setFavorited(true);
                toaster('Library added to favorites', ToastState.success);
            } else {
                toaster(response.message, ToastState.error);
            }
        } catch (error) {
            toaster('Error updating favorite status', ToastState.error);
            console.error(error);
        } finally {
            favoriteModalRef.current?.close();
            reset();
        }
    };

    const favoriteIcon = favorited ? (
        <StarIcon
            className="w-6 text-primary-yellow cursor-pointer"
            onClick={toggleFavorite}
        />
    ) : (
        <StarIconOutline
            className="w-6 text-header-text cursor-pointer"
            onClick={toggleFavorite}
        />
    );

    return (
        <div>
            <div className="px-8 pb-4 flex justify-between items-center">
                <div className="flex items-center">
                    <h1 className="mr-4">Library Viewer</h1>
                    {favoriteIcon}
                </div>
            </div>
            <div className="w-full pt-4 justify-center">
                {isLoading ? (
                    <div className="flex h-screen gap-4 justify-center content-center">
                        <span className="my-auto loading loading-spinner loading-lg"></span>
                        <p className="my-auto text-lg">Loading...</p>
                    </div>
                ) : src !== '' ? (
                    <iframe
                        sandbox="allow-scripts allow-same-origin"
                        className="w-full h-screen pt-4"
                        id="library-viewer"
                        src={src}
                    />
                ) : (
                    error && <Error />
                )}
            </div>

            <Modal
                ref={favoriteModalRef}
                type={ModalType.Add}
                item="Favorite"
                form={
                    <form
                        onSubmit={(e) =>
                            void handleSubmit(handleFavoriteSubmit)(e)
                        }
                        className="flex flex-col space-y-4"
                    >
                        <TextInput
                            label="Favorite Name"
                            interfaceRef="favoriteName"
                            required={true}
                            length={100}
                            errors={errors}
                            register={register}
                        />
                        <div className="flex justify-end space-x-4">
                            <SubmitButton />
                        </div>
                        <CloseX
                            close={() => favoriteModalRef.current?.close()}
                        />
                    </form>
                }
            />
        </div>
    );
}
