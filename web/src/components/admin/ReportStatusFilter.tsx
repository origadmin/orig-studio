/**
 * ReportStatusFilter molecule - dropdown filter for report status.
 */
import React from 'react';
import { useTranslation } from 'react-i18next';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Filter } from 'lucide-react';

interface ReportStatusFilterProps {
  value: string;
  onChange: (value: string) => void;
}

const REPORT_STATUS_OPTIONS = [
  { value: 'all', label: 'All Comments' },
  { value: 'reported', label: 'Has Reports' },
  { value: 'pending_reports', label: 'Pending Review' },
  { value: 'reviewed_reports', label: 'Review Completed' },
  { value: 'no_reports', label: 'No Reports' },
];

export const ReportStatusFilter: React.FC<ReportStatusFilterProps> = React.memo(
  ({ value, onChange }) => {
    const { t } = useTranslation();

    return (
      <Select value={value} onValueChange={onChange}>
        <SelectTrigger className="w-[160px] h-8 rounded-btn-sm focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0">
          <div className="flex items-center gap-2">
            <Filter className="h-4 w-4" />
            {value === 'all' ? (
              <span className="text-muted-foreground">Reports</span>
            ) : (
              <SelectValue placeholder="Reports" />
            )}
          </div>
        </SelectTrigger>
        <SelectContent>
          {REPORT_STATUS_OPTIONS.map(option => (
            <SelectItem key={option.value} value={option.value}>
              {option.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    );
  }
);

ReportStatusFilter.displayName = 'ReportStatusFilter';
