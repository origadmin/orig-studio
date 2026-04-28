import React, {useState, useEffect, useRef} from 'react';
import {useTranslation} from 'react-i18next';
import {Button} from '@/components/ui/button';
import {CheckCircle, Bell, BellOff} from 'lucide-react';

interface NotificationBellProps {
    isSubscribed: boolean;
    currentSetting?: string;
    onSettingChange?: (setting: string) => Promise<void> | void;
}

type NotificationSetting = 'all' | 'personalized' | 'none';

const NotificationBell: React.FC<NotificationBellProps> = ({
    isSubscribed,
    currentSetting = 'all',
    onSettingChange,
}) => {
    const {t} = useTranslation();
    const [isOpen, setIsOpen] = useState(false);
    const [setting, setSetting] = useState<NotificationSetting>(
        currentSetting as NotificationSetting
    );
    const [changing, setChanging] = useState(false);
    const menuRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        setSetting(currentSetting as NotificationSetting);
    }, [currentSetting]);

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
                setIsOpen(false);
            }
        };

        if (isOpen) {
            document.addEventListener('mousedown', handleClickOutside);
        }

        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, [isOpen]);

    const options: Array<{
        value: NotificationSetting;
        labelKey: string;
        descKey: string;
    }> = [
        {value: 'all', labelKey: 'channel.notifyAll', descKey: 'channel.notifyAllDesc'},
        {value: 'personalized', labelKey: 'channel.notifyPersonalized', descKey: 'channel.notifyPersonalizedDesc'},
        {value: 'none', labelKey: 'channel.notifyNone', descKey: 'channel.notifyNoneDesc'},
    ];

    const handleSettingChange = async (newSetting: NotificationSetting) => {
        if (changing || newSetting === setting) {
            setIsOpen(false);
            return;
        }

        try {
            setChanging(true);
            setSetting(newSetting);
            await onSettingChange?.(newSetting);
            setIsOpen(false);
        } catch (error) {
            console.error('Failed to change notification setting:', error);
            setSetting(setting);
        } finally {
            setChanging(false);
        }
    };

    return (
        <div className="relative" ref={menuRef}>
            <Button
                variant="outline"
                size="icon"
                onClick={() => isSubscribed && setIsOpen(!isOpen)}
                className={!isSubscribed ? 'opacity-50 cursor-not-allowed' : ''}
                disabled={!isSubscribed}
                title={isSubscribed ? t('channel.notificationSettings') : t('channel.pleaseLogin') + ' → ' + t('channel.subscribe')}
            >
                {isSubscribed ? (
                    <Bell
                        className={`w-4 h-4 ${
                            setting !== 'none'
                                ? 'fill-current text-primary'
                                : ''
                        }`}
                    />
                ) : (
                    <BellOff className="w-4 h-4"/>
                )}
            </Button>

            {isOpen && isSubscribed && (
                <div
                    className="absolute right-0 top-full mt-2 w-64 bg-popover rounded-lg border shadow-lg z-50 animate-in fade-in slide-in-from-top-2 duration-200"
                    role="menu"
                    aria-label={t('channel.notificationSettings')}
                >
                    <div className="p-3 space-y-1">
                        <p className="text-sm font-medium mb-2">
                            {t('channel.notificationSettings')}
                        </p>
                        {options.map((option) => (
                            <label
                                key={option.value}
                                className={`flex items-start gap-3 p-2 rounded-md hover:bg-accent cursor-pointer transition-colors ${
                                    changing ? 'pointer-events-none opacity-60' : ''
                                }`}
                                onClick={() =>
                                    handleSettingChange(option.value)
                                }
                                role="menuitemradio"
                                aria-checked={setting === option.value}
                            >
                                <div
                                    className={`mt-0.5 w-4 h-4 rounded-full border-2 flex-shrink-0 transition-colors ${
                                        setting === option.value
                                            ? 'border-primary bg-primary'
                                            : 'border-muted-foreground/30'
                                    }`}
                                >
                                    {setting === option.value && (
                                        <CheckCircle className="w-3 h-3 text-primary-foreground m-0.5"/>
                                    )}
                                </div>
                                <div className="min-w-0">
                                    <p className="text-sm font-medium">
                                        {t(option.labelKey)}
                                    </p>
                                    <p className="text-xs text-muted-foreground">
                                        {t(option.descKey)}
                                    </p>
                                </div>
                            </label>
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
};

export default NotificationBell;
