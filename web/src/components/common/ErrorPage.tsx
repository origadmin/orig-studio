import React from 'react';
import {Link} from '@tanstack/react-router';
import {Button} from '@/components/ui/button';
import {useTranslation} from 'react-i18next';

interface ErrorPageProps {
    statusCode?: number;
    title?: string;
    message?: string;
    showBackButton?: boolean;
    backButtonText?: string;
    backButtonPath?: string;
}

const ErrorPage: React.FC<ErrorPageProps> = ({
                                                 statusCode,
                                                 title,
                                                 message,
                                                 showBackButton = true,
                                                 backButtonText,
                                                 backButtonPath = '/',
                                             }) => {
    const {t} = useTranslation();

    const defaultTitle = statusCode === 404 ? t('error.404Title') : t('error.genericTitle');
    const defaultMessage = statusCode === 404 ? t('error.404Message') : t('error.genericMessage');
    const defaultBackButtonText = t('error.backToHome');

    return (
        <div className="flex flex-col items-center justify-center min-h-[80vh] px-4 text-center">
            {statusCode && (
                <div className="text-9xl font-bold text-emerald-600 mb-6">{statusCode}</div>
            )}
            <h1 className="text-3xl font-bold mb-4">{title || defaultTitle}</h1>
            <p className="text-lg text-gray-600 dark:text-gray-300 mb-8 max-w-2xl">
                {message || defaultMessage}
            </p>
            {showBackButton && (
                <Link to={backButtonPath}>
                    <Button size="lg" className="px-6 py-3">
                        {backButtonText || defaultBackButtonText}
                    </Button>
                </Link>
            )}
        </div>
    );
};

export default ErrorPage;
