import React, {useState, useRef, useCallback} from 'react';
import {ChevronLeft, ChevronRight} from 'lucide-react';
import {Button} from '@/components/ui/button';

interface HorizontalScrollProps {
    children: React.ReactNode;
    className?: string;
    itemsPerPage?: number;
}

const HorizontalScroll: React.FC<HorizontalScrollProps> = ({children, className = '', itemsPerPage = 4}) => {
    const [currentPage, setCurrentPage] = useState(0);
    const containerRef = useRef<HTMLDivElement>(null);
    const childrenArray = React.Children.toArray(children);
    const totalPages = Math.ceil(childrenArray.length / itemsPerPage);

    const goToPage = useCallback((page: number) => {
        setCurrentPage(page);
        if (containerRef.current) {
            containerRef.current.scrollTo({
                left: page * containerRef.current.clientWidth,
                behavior: 'smooth'
            });
        }
    }, []);

    const nextPage = useCallback(() => {
        if (currentPage < totalPages - 1) {
            goToPage(currentPage + 1);
        }
    }, [currentPage, totalPages, goToPage]);

    const prevPage = useCallback(() => {
        if (currentPage > 0) {
            goToPage(currentPage - 1);
        }
    }, [currentPage, goToPage]);

    return (
        <div className={`relative ${className}`}>
            <div
                ref={containerRef}
                className="flex gap-4 overflow-x-auto pb-4 scrollbar-hide"
                style={{
                    scrollbarWidth: 'none',
                    msOverflowStyle: 'none',
                    overflowY: 'hidden',
                    whiteSpace: 'nowrap'
                }}
            >
                <div className="flex gap-4">
                    {children}
                </div>
            </div>
            {totalPages > 1 && (
                <>
                    <Button
                        variant="ghost"
                        size="icon"
                        className={`absolute left-0 top-1/2 -translate-y-1/2 bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm shadow-md z-10 ${currentPage === 0 ? 'opacity-50 cursor-not-allowed' : ''}`}
                        onClick={prevPage}
                        disabled={currentPage === 0}
                    >
                        <ChevronLeft size={20}/>
                    </Button>
                    <Button
                        variant="ghost"
                        size="icon"
                        className={`absolute right-0 top-1/2 -translate-y-1/2 bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm shadow-md z-10 ${currentPage === totalPages - 1 ? 'opacity-50 cursor-not-allowed' : ''}`}
                        onClick={nextPage}
                        disabled={currentPage === totalPages - 1}
                    >
                        <ChevronRight size={20}/>
                    </Button>
                </>
            )}
        </div>
    );
};

export default HorizontalScroll;