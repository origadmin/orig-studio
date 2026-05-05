/**
 * CommentStatsCards organism - displays statistics cards for comments.
 */
import React from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardContent } from '@/components/ui/card';
import { MessageCircle, CheckCircle, Clock, ShieldOff, AlertTriangle } from 'lucide-react';
import type { CommentStats } from '@/lib/api/comment';

interface CommentStatsCardsProps {
  stats: CommentStats | null;
  loading?: boolean;
}

export const CommentStatsCards: React.FC<CommentStatsCardsProps> = React.memo(
  ({ stats, loading }) => {
    const { t } = useTranslation();

    if (loading || !stats) {
      return null;
    }

    const cards = [
      {
        label: t('admin.totalComments') || 'Total',
        value: stats.total,
        icon: MessageCircle,
        color: 'text-info',
        bgColor: 'bg-info',
      },
      {
        label: t('admin.approved') || 'Approved',
        value: stats.approved,
        icon: CheckCircle,
        color: 'text-success',
        bgColor: 'bg-success',
      },
      {
        label: t('admin.pending') || 'Pending',
        value: stats.pending,
        icon: Clock,
        color: 'text-yellow-600',
        bgColor: 'bg-warning',
      },
      {
        label: 'Blocked',
        value: stats.blocked,
        icon: ShieldOff,
        color: 'text-destructive',
        bgColor: 'bg-destructive',
      },
      {
        label: 'Pending Reports',
        value: stats.reported_pending,
        icon: AlertTriangle,
        color: 'text-orange-600',
        bgColor: 'bg-orange-500',
      },
    ];

    return (
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
        {cards.map(card => (
          <Card key={card.label} className="relative overflow-hidden shadow-sm border-none ring-1 ring-border">
            <CardContent className="pt-6">
              <div className="flex items-center gap-2">
                <card.icon className={`h-5 w-5 ${card.color}`} />
                <div>
                  <div className={`text-2xl font-bold ${card.color}`}>{card.value}</div>
                  <p className="text-sm text-muted-foreground">{card.label}</p>
                </div>
              </div>
            </CardContent>
            <div className={`absolute bottom-0 left-0 h-1 ${card.bgColor} w-full opacity-10`} />
          </Card>
        ))}
      </div>
    );
  }
);

CommentStatsCards.displayName = 'CommentStatsCards';
