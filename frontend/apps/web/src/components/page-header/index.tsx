import { Fragment } from "react";
import type { JSX } from "react";
import { Link } from "@tanstack/react-router";
import { Breadcrumb } from "@pdv/ui-kit/components/breadcrumb";
import type {
  PageHeaderBreadcrumb,
  PageHeaderProps,
} from "../../interfaces/page-header-props.interface";

export function PageHeader({
  breadcrumbs,
  title,
  description,
  action,
  actionPlacement = "side",
}: PageHeaderProps): JSX.Element {
  return (
    <>
      <div
        className={`${actionPlacement === "below" ? "mb-5" : "mb-8"} flex flex-col justify-between gap-5 md:flex-row md:items-end`}
      >
        <div>
          <Breadcrumb className="mb-4">
            <Breadcrumb.List className="text-[11px] uppercase tracking-[0.16em]">
              {breadcrumbs.map(
                (breadcrumb: PageHeaderBreadcrumb, index: number) => (
                  <Fragment key={breadcrumb.label}>
                    {index > 0 && <Breadcrumb.Separator />}
                    <Breadcrumb.Item>
                      {breadcrumb.to && index < breadcrumbs.length - 1 ? (
                        <Breadcrumb.Link
                          render={<Link to={breadcrumb.to} />}
                          className="font-semibold text-(--ink-soft)"
                        >
                          {breadcrumb.label}
                        </Breadcrumb.Link>
                      ) : index === breadcrumbs.length - 1 ? (
                        <Breadcrumb.Page className="font-semibold text-(--ink)">
                          {breadcrumb.label}
                        </Breadcrumb.Page>
                      ) : (
                        <span className="font-semibold text-(--ink-soft)">
                          {breadcrumb.label}
                        </span>
                      )}
                    </Breadcrumb.Item>
                  </Fragment>
                ),
              )}
            </Breadcrumb.List>
          </Breadcrumb>
          <h1>{title}</h1>
          <p className="mt-3 max-w-xl text-sm leading-relaxed text-(--ink-soft)">
            {description}
          </p>
        </div>
        {action && actionPlacement === "side" && (
          <div className="shrink-0">{action}</div>
        )}
      </div>
      {action && actionPlacement === "below" && (
        <div className="mb-8">{action}</div>
      )}
    </>
  );
}
