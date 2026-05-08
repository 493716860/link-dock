export type WorkflowStatus = 'none' | 'read_later' | 'unorganized';

export interface Category {
  id: string;
  name: string;
  iconName: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface SeedCategory {
  id: string;
  name: string;
  iconName: string;
}

export interface SeedSite {
  id: string;
  categoryId: string;
  name: string;
  url: string;
  description: string;
  icon?: string;
}

export interface SeedConfig {
  categories: SeedCategory[];
  sites: SeedSite[];
}

export interface Site {
  id: string;
  name: string;
  url: string;
  description: string;
  categoryId: string;
  sortOrder?: number;
  icon?: string;
  isPublic?: boolean;
  tagsText?: string;
  isFavorite?: boolean;
  workflowStatus?: WorkflowStatus;
  visitCount?: number;
  lastVisitedAt?: string;
  createdAt?: string;
  updatedAt?: string;
}
