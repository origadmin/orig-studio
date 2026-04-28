import {Button} from '@/components/ui/button';
import {ChevronLeft, ChevronRight} from 'lucide-react';

interface TablePaginationProps {
    page: number;
    pageSize: number;
    total: number;
    onPageChange: (page: number) => void;
}

export function TablePagination({page, pageSize, total, onPageChange}: TablePaginationProps) {
    const totalPages = Math.ceil(total / pageSize);

    if (total <= pageSize) {
        return null;
    }

    return (
        <div className="flex items-center justify-between pt-1 text-xs text-muted-foreground">
            <span className="tabular-nums">
                Page {page} of {totalPages} · {total} total
            </span>
            <div className="flex gap-1.5">
                <Button
                    variant="outline"
                    size="sm"
                    className="h-9 px-3"
                    disabled={page <= 1}
                    onClick={() => onPageChange(page - 1)}
                >
                    <ChevronLeft className="h-4 w-4 mr-1"/>
                    Previous
                </Button>
                <Button
                    variant="outline"
                    size="sm"
                    className="h-9 px-3"
                    disabled={page >= totalPages}
                    onClick={() => onPageChange(page + 1)}
                >
                    Next
                    <ChevronRight className="h-4 w-4 ml-1"/>
                </Button>
            </div>
        </div>
    );
}
