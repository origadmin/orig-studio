/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import React from 'react';
import {Link} from '@tanstack/react-router';
import {Globe, Heart, Mail, Video, MessageCircle} from 'lucide-react';

const Footer = () => {
    return (
        <footer className="bg-muted border-t border-border py-8">
            <div className="container mx-auto px-4">
                <div className="grid grid-cols-2 md:grid-cols-4 gap-8 mb-8">
                    <div className="space-y-6">
                        <Link to="/" className="flex items-center space-x-2">
                            <img src="/logo.svg" alt="OrigCMS" className="h-8 w-8" />
                            <span className="text-xl font-black text-foreground tracking-tighter">OrigCMS</span>
                        </Link>
                        <p className="text-sm text-muted-foreground font-medium leading-relaxed">
                            Next-generation video platform powered by Go microservices.
                        </p>
                        <div className="flex items-center space-x-4">
                            <SocialIcon icon={<Globe size={18}/>}/>
                            <SocialIcon icon={<Mail size={18}/>}/>
                            <SocialIcon icon={<Video size={18}/>}/>
                            <SocialIcon icon={<MessageCircle size={18}/>}/>
                        </div>
                    </div>

                    <FooterSection title="Platform" links={[
                        {label: 'Explore', to: '/'},
                        {label: 'Trending', to: '/trending'},
                        {label: 'Categories', to: '/categories'},
                        {label: 'Channels', to: '/c/1'},
                    ]}/>

                    <FooterSection title="Create" links={[
                        {label: 'Upload Video', to: '/me/upload'},
                        {label: 'Start Streaming', to: '/live'},
                    ]}/>

                    <FooterSection title="Account" links={[
                        {label: 'My Profile', to: '/u/1'},
                        {label: 'My Favorites', to: '/me/favorites'},
                        {label: 'Notifications', to: '/me/notifications'},
                        {label: 'Sign In', to: '/auth/signin'},
                    ]}/>
                </div>

                <div
                    className="pt-8 border-t border-border flex flex-col md:flex-row justify-between items-center gap-4 text-[11px] font-bold text-muted-foreground uppercase tracking-widest">
                    <div className="flex items-center gap-4">
                        <Link to="/privacy" className="hover:text-info transition-colors">Privacy Policy</Link>
                        <Link to="/terms" className="hover:text-info transition-colors">Terms of Service</Link>
                        <Link to="/cookies" className="hover:text-info transition-colors">Cookie Policy</Link>
                    </div>
                    <p className="flex items-center gap-1.5">
                        Made with <Heart size={12} className="text-destructive fill-destructive"/> by
                        <span className="text-foreground">OrigAdmin Team</span> © 2024
                    </p>
                </div>
            </div>
        </footer>
    );
};

const FooterSection = ({title, links}: {
    title: string;
    links: { label: string; to: string; search?: Record<string, any> }[]
}) => (
    <div className="space-y-6">
        <h4 className="text-sm font-black text-foreground uppercase tracking-widest">{title}</h4>
        <ul className="space-y-4">
            {links.map((link) => (
                <li key={link.label}>
                    <Link to={link.to} search={link.search}
                          className="text-sm text-muted-foreground font-bold hover:text-info transition-all hover:translate-x-1 inline-block">
                        {link.label}
                    </Link>
                </li>
            ))}
        </ul>
    </div>
);

const SocialIcon = ({icon}: any) => (
    <button
        className="w-9 h-9 bg-card border border-border rounded-xl flex items-center justify-center text-muted-foreground hover:text-info hover:border-brand-muted hover:bg-brand-muted transition-all shadow-sm hover:shadow-md active:scale-95">
        {icon}
    </button>
);

export default Footer;
