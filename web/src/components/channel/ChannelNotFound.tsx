import React from 'react';
import {useTranslation} from 'react-i18next';
import {useNavigate} from '@tanstack/react-router';
import {Button} from '@/components/ui/button';
import {Home, ArrowLeft, Search} from 'lucide-react';

interface ChannelNotFoundProps {
    message?: string;
    onBack?: () => void;
}

const ChannelNotFound: React.FC<ChannelNotFoundProps> = ({
    message,
    onBack,
}) => {
    const {t} = useTranslation();
    const navigate = useNavigate();

    return (
        <div className="min-h-[60vh] flex items-center justify-center bg-background">
            <div className="text-center space-y-6 px-4 max-w-md">
                <div className="w-24 h-24 mx-auto rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
                    <Search className="w-12 h-12 text-muted-foreground"/>
                </div>

                <div className="space-y-2">
                    <h2 className="text-2xl font-bold text-foreground">
                        {t('channel.notFoundTitle')}
                    </h2>
                    <p className="text-muted-foreground">
                        {message || t('channel.notFoundDesc')}
                    </p>
                </div>

                <div className="flex flex-col sm:flex-row items-center justify-center gap-3 pt-4">
                    <Button
                        variant="outline"
                        onClick={onBack || (() => navigate({to: '/'}))}
                        className="gap-2"
                    >
                        <ArrowLeft className="w-4 h-4"/>
                        {t('channel.backToHome')}
                    </Button>
                    <Button
                        onClick={() => navigate({to: '/'})}
                        className="gap-2"
                    >
                        <Home className="w-4 h-4"/>
                        {t('channel.exploreChannels')}
                    </Button>
                </div>

                <div className="pt-4 border-t border-border">
                    <p className="text-sm text-muted-foreground">
                        {t('channel.checkUrl')}
                    </p>
                    <ul className="text-sm text-muted-foreground mt-2 space-y-1">
                        <li>• {t('channel.checkUrl1')}</li>
                        <li>• {t('channel.checkUrl2')}</li>
                        <li>• {t('channel.checkUrl3')}</li>
                    </ul>
                </div>
            </div>
        </div>
    );
};

export default ChannelNotFound;
