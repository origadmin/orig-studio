import React, {useState} from 'react';
import {useTranslation} from 'react-i18next';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import {Button} from '@/components/ui/button';
import {Textarea} from '@/components/ui/textarea';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/components/ui/select';
import {Loader2, Flag} from 'lucide-react';

interface ReportDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    targetId: string;
    targetType: 'comment' | 'media';
    onSubmit: (data: { reason: string; description?: string }) => Promise<void>;
}

const REPORT_REASONS = [
    {value: 'SPAM', labelKey: 'report.reasonSpam', defaultLabel: 'Spam'},
    {value: 'HARASSMENT', labelKey: 'report.reasonHarassment', defaultLabel: 'Harassment'},
    {value: 'INAPPROPRIATE', labelKey: 'report.reasonInappropriate', defaultLabel: 'Inappropriate Content'},
    {value: 'OTHER', labelKey: 'report.reasonOther', defaultLabel: 'Other'},
];

const ReportDialog: React.FC<ReportDialogProps> = React.memo(
    ({open, onOpenChange, targetId, targetType, onSubmit}) => {
        const {t} = useTranslation();
        const [reason, setReason] = useState('');
        const [description, setDescription] = useState('');
        const [submitting, setSubmitting] = useState(false);
        const [error, setError] = useState<string | null>(null);

        const handleSubmit = async () => {
            if (!reason) return;
            try {
                setSubmitting(true);
                setError(null);
                await onSubmit({reason, description: description || undefined});
                setReason('');
                setDescription('');
                onOpenChange(false);
            } catch (err: any) {
                if (err?.message?.includes('already reported')) {
                    setError(t('report.alreadyReported') || 'You have already reported this content');
                } else {
                    setError(err?.message || t('report.submitFailed') || 'Failed to submit report');
                }
            } finally {
                setSubmitting(false);
            }
        };

        const handleOpenChange = (newOpen: boolean) => {
            if (!newOpen) {
                setReason('');
                setDescription('');
                setError(null);
            }
            onOpenChange(newOpen);
        };

        const titleKey = targetType === 'media' ? 'report.reportVideo' : 'report.reportComment';
        const descKey = targetType === 'media' ? 'report.reportVideoDesc' : 'report.reportCommentDesc';

        return (
            <Dialog open={open} onOpenChange={handleOpenChange}>
                <DialogContent className="sm:max-w-[425px]">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            <Flag className="w-5 h-5 text-destructive"/>
                            {t(titleKey) || (targetType === 'media' ? 'Report Video' : 'Report Comment')}
                        </DialogTitle>
                        <DialogDescription>
                            {t(descKey) || 'Submit a report. Our team will review it.'}
                        </DialogDescription>
                    </DialogHeader>

                    <div className="space-y-4 py-4">
                        <div className="space-y-2">
                            <label className="text-sm font-medium">
                                {t('report.reason') || 'Reason'} *
                            </label>
                            <Select value={reason} onValueChange={setReason}>
                                <SelectTrigger>
                                    <SelectValue placeholder={t('report.selectReason') || 'Select a reason'}/>
                                </SelectTrigger>
                                <SelectContent>
                                    {REPORT_REASONS.map(r => (
                                        <SelectItem key={r.value} value={r.value}>
                                            {t(r.labelKey) || r.defaultLabel}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>

                        <div className="space-y-2">
                            <label className="text-sm font-medium">
                                {t('report.description') || 'Description'} ({t('common.optional') || 'optional'})
                            </label>
                            <Textarea
                                value={description}
                                onChange={(e) => setDescription(e.target.value)}
                                placeholder={t('report.descriptionPlaceholder') || 'Add more details...'}
                                rows={3}
                            />
                        </div>

                        {error && (
                            <p className="text-sm text-destructive">{error}</p>
                        )}
                    </div>

                    <DialogFooter>
                        <Button
                            variant="outline"
                            onClick={() => handleOpenChange(false)}
                            disabled={submitting}
                        >
                            {t('common.cancel') || 'Cancel'}
                        </Button>
                        <Button
                            onClick={handleSubmit}
                            disabled={!reason || submitting}
                            variant="destructive"
                        >
                            {submitting ? (
                                <>
                                    <Loader2 className="mr-2 h-4 w-4 animate-spin"/>
                                    {t('report.submitting') || 'Submitting...'}
                                </>
                            ) : (
                                t('report.submit') || 'Submit Report'
                            )}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        );
    }
);

ReportDialog.displayName = 'ReportDialog';

export default ReportDialog;
