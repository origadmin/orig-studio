/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import React, { ReactNode } from 'react';

interface ButtonGroupProps {
  children: ReactNode;
  className?: string;
  orientation?: 'horizontal' | 'vertical';
}

export const ButtonGroup: React.FC<ButtonGroupProps> = ({
  children,
  className = '',
  orientation = 'horizontal',
}) => {
  const baseClasses = 'flex gap-2';
  const orientationClasses = orientation === 'horizontal' ? 'flex-row' : 'flex-col';
  
  return (
    <div className={`${baseClasses} ${orientationClasses} ${className}`}>
      {children}
    </div>
  );
};
