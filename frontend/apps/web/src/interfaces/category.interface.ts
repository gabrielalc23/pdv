export interface CategoryResponse {
  id: string;
  name: string;
  slug: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CategoryListResponse {
  data: CategoryResponse[];
}

export interface ListCategoriesParams {
  search?: string;
  activeOnly?: boolean;
}
