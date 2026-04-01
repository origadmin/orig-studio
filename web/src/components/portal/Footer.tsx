/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import React from 'react';
import {Link} from '@tanstack/react-router';
import {Globe, Heart, Mail, Video, MessageCircle} from 'lucide-react';

const Footer = () => {
    return (
        <footer className="bg-slate-50 border-t border-slate-100 py-8">
            <div className="container mx-auto px-4">
                <div className="grid grid-cols-2 md:grid-cols-4 gap-8 mb-8">
                    <div className="space-y-6">
                        <Link to="/" className="flex items-center space-x-2">
                            <div
                                className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center font-bold text-white shadow-lg">O
                            </div>
                            <span className="text-xl font-black text-slate-900 tracking-tighter">ORIGCMS</span>
                        </Link>
                        <p className="text-sm text-slate-500 font-medium leading-relaxed">
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
                    className="pt-8 border-t border-slate-200 flex flex-col md:flex-row justify-between items-center gap-4 text-[11px] font-bold text-slate-400 uppercase tracking-widest">
                    <div className="flex items-center gap-4">
                        <Link to="/privacy" className="hover:text-blue-600 transition-colors">Privacy Policy</Link>
                        <Link to="/terms" className="hover:text-blue-600 transition-colors">Terms of Service</Link>
                        <Link to="/cookies" className="hover:text-blue-600 transition-colors">Cookie Policy</Link>
                    </div>
                    <p className="flex items-center gap-1.5">
                        Made with <Heart size={12} className="text-rose-500 fill-rose-500"/> by
                        <span className="text-slate-900">OrigAdmin Team</span> © 2024
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
        <h4 className="text-sm font-black text-slate-900 uppercase tracking-widest">{title}</h4>
        <ul className="space-y-4">
            {links.map((link) => (
                <li key={link.label}>
                    <Link to={link.to} search={link.search}
                          className="text-sm text-slate-500 font-bold hover:text-blue-600 transition-all hover:translate-x-1 inline-block">
                        {link.label}
                    </Link>
                </li>
            ))}
        </ul>
    </div>
);

const SocialIcon = ({icon}: any) => (
    <button
        className="w-9 h-9 bg-white border border-slate-200 rounded-xl flex items-center justify-center text-slate-400 hover:text-blue-600 hover:border-blue-200 hover:bg-blue-50 transition-all shadow-sm hover:shadow-md active:scale-95">
        {icon}
    </button>
);

export default Footer;
