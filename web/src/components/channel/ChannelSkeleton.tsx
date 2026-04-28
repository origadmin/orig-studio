import React from 'react';

const ChannelSkeleton: React.FC = () => {
    return (
        <div className="channel-page min-h-screen bg-background animate-pulse">
            <div className="max-w-[1920px] mx-auto">
                {/* Banner skeleton */}
                <div className="w-full h-[150px] sm:h-[200px] md:h-[250px] bg-muted"/>

                {/* Channel info bar skeleton */}
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="flex flex-col sm:flex-row items-start sm:items-end gap-4 sm:gap-6 -mt-10 sm:-mt-14 relative z-10 pb-4">
                        {/* Avatar skeleton */}
                        <div className="w-20 h-20 sm:w-28 sm:h-28 md:w-[120px] md:h-[120px] rounded-full bg-muted border-4 border-background flex-shrink-0"/>
                        {/* Info skeleton */}
                        <div className="flex-1 pt-2 sm:pt-4 space-y-3">
                            <div className="h-8 w-64 bg-muted rounded"/>
                            <div className="h-4 w-96 bg-muted rounded"/>
                            <div className="h-4 w-full max-w-2xl bg-muted rounded"/>
                        </div>
                        {/* Buttons skeleton */}
                        <div className="flex gap-2 pt-2 sm:pt-4 flex-shrink-0">
                            <div className="h-9 w-28 bg-muted rounded-lg"/>
                            <div className="h-9 w-9 bg-muted rounded-full"/>
                            <div className="h-9 w-9 bg-muted rounded-full"/>
                        </div>
                    </div>
                </div>

                {/* Nav skeleton */}
                <div className="border-b border-border bg-background">
                    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                        <nav className="flex gap-8">
                            {[1, 2, 3, 4, 5].map((i) => (
                                <div
                                    key={i}
                                    className="py-4 px-4 h-8 bg-muted rounded-t"
                                />
                            ))}
                        </nav>
                    </div>
                </div>

                {/* Content skeleton */}
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
                    <div className="flex flex-col lg:flex-row gap-6">
                        <div className="flex-1 min-w-0">
                            <div className="min-h-[400px] space-y-4">
                                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                                    {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
                                        <div key={i} className="space-y-2">
                                            <div className="aspect-video bg-muted rounded-lg"/>
                                            <div className="space-y-2">
                                                <div className="h-4 w-full bg-muted rounded"/>
                                                <div className="h-4 w-3/4 bg-muted rounded"/>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                        {/* Sidebar skeleton */}
                        <div className="hidden lg:block w-80 flex-shrink-0">
                            <div className="space-y-4">
                                {[1, 2, 3].map((i) => (
                                    <div key={i} className="flex items-center gap-3">
                                        <div className="w-10 h-10 bg-muted rounded-full"/>
                                        <div className="flex-1 space-y-2">
                                            <div className="h-4 w-3/4 bg-muted rounded"/>
                                            <div className="h-3 w-1/2 bg-muted rounded"/>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default ChannelSkeleton;
