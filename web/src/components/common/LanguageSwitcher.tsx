/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * LanguageSwitcher: Dropdown menu for switching between supported languages.
 */

import React from 'react';
import {useTranslation} from 'react-i18next';
import {Globe, Check} from 'lucide-react';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

interface LanguageOption {
    code: string;
    label: string;
    nativeLabel: string;
}

const LANGUAGES: LanguageOption[] = [
    {code: 'zh', label: 'Chinese', nativeLabel: '中文'},
    {code: 'en', label: 'English', nativeLabel: 'English'},
    {code: 'ja', label: 'Japanese', nativeLabel: '日本語'},
];

interface LanguageSwitcherProps {
    className?: string;
    buttonClassName?: string;
    variant?: 'default' | 'compact';
}

const LanguageSwitcher: React.FC<LanguageSwitcherProps> = ({
    className,
    buttonClassName,
    variant = 'default',
}) => {
    const {i18n} = useTranslation();
    const currentLanguage = i18n.language;

    const handleLanguageChange = (langCode: string) => {
        i18n.changeLanguage(langCode);
    };

    const currentLang = LANGUAGES.find(l => l.code === currentLanguage) || LANGUAGES[0];

    return (
        <DropdownMenu>
            <DropdownMenuTrigger
                className={`inline-flex items-center justify-center rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-brand focus:ring-offset-2 ${variant === 'compact' ? 'w-8 h-8' : 'w-10 h-10'} hover:bg-accent ${buttonClassName || ''}`}
                aria-label={currentLang.nativeLabel}
            >
                <Globe size={variant === 'compact' ? 16 : 18} className="text-brand"/>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className={className}>
                {LANGUAGES.map((lang) => (
                    <DropdownMenuItem
                        key={lang.code}
                        onClick={() => handleLanguageChange(lang.code)}
                        className="flex items-center justify-between gap-4 cursor-pointer"
                    >
                        <span className="flex items-center gap-2">
                            <span className="text-sm font-medium">{lang.nativeLabel}</span>
                            <span className="text-xs text-muted-foreground">{lang.label}</span>
                        </span>
                        {currentLanguage === lang.code && (
                            <Check size={16} className="text-brand"/>
                        )}
                    </DropdownMenuItem>
                ))}
            </DropdownMenuContent>
        </DropdownMenu>
    );
};

export default LanguageSwitcher;
