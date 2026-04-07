import React, {useState, useEffect} from 'react';
import {Globe} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger} from '@/components/ui/dropdown-menu';
import {useTranslation} from 'react-i18next';
import {subtitleApi, type Subtitle} from '@/lib/api/subtitle';

interface SubtitleSelectorProps {
    mediaId: string;
    onSubtitleChange: (subtitle: Subtitle | null) => void;
}

const SubtitleSelector: React.FC<SubtitleSelectorProps> = ({mediaId, onSubtitleChange}) => {
    const {t} = useTranslation();
    const [subtitles, setSubtitles] = useState<Subtitle[]>([]);
    const [loading, setLoading] = useState(true);
    const [selectedSubtitle, setSelectedSubtitle] = useState<Subtitle | null>(null);

    useEffect(() => {
        fetchSubtitles();
    }, [mediaId]);

    const fetchSubtitles = async () => {
        try {
            setLoading(true);
            const response = await subtitleApi.getByMediaId(mediaId);
            setSubtitles(response || []);
        } catch (err) {
            console.error('Failed to fetch subtitles:', err);
        } finally {
            setLoading(false);
        }
    };

    const handleSubtitleSelect = (subtitle: Subtitle | null) => {
        setSelectedSubtitle(subtitle);
        onSubtitleChange(subtitle);
    };

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button variant="secondary" size="sm"
                        className="bg-black/70 hover:bg-black/90 text-white border-white/20 backdrop-blur-md text-xs gap-1.5 h-7 px-2">
                    <Globe size={12}/>
                    {selectedSubtitle ? selectedSubtitle.language_name : t('watch.subtitles')}
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-48 bg-black/90 text-white border-white/20 backdrop-blur-md">
                <DropdownMenuItem onClick={() => handleSubtitleSelect(null)} className="text-white hover:bg-white/10">
                    {t('watch.off')}
                </DropdownMenuItem>
                {loading ? (
                    <DropdownMenuItem disabled className="text-white/50">
                        {t('common.loading')}
                    </DropdownMenuItem>
                ) : subtitles.length === 0 ? (
                    <DropdownMenuItem disabled className="text-white/50">
                        {t('watch.noSubtitles')}
                    </DropdownMenuItem>
                ) : (
                    subtitles.map((subtitle) => (
                        <DropdownMenuItem
                            key={subtitle.id}
                            onClick={() => handleSubtitleSelect(subtitle)}
                            className={`text-white ${selectedSubtitle?.id === subtitle.id ? 'bg-white/10' : 'hover:bg-white/10'}`}
                        >
                            {subtitle.language_name}
                        </DropdownMenuItem>
                    ))
                )}
            </DropdownMenuContent>
        </DropdownMenu>
    );
};

export default SubtitleSelector;
