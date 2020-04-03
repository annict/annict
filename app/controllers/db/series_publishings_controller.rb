# frozen_string_literal: true

module Db
  class SeriesPublishingsController < Db::ApplicationController
    before_action :authenticate_user!

    def create
      @series = Series.without_deleted.unpublished.find(params[:id])
      authorize_db_resource_publishing @series

      @series.publish

      redirect_back(
        fallback_location: db_series_list_path,
        notice: t("messages._common.published")
      )
    end

    def destroy
      @series = Series.without_deleted.published.find(params[:id])
      authorize_db_resource_publishing @series

      @series.unpublish

      redirect_back(
        fallback_location: db_series_list_path,
        notice: t("messages._common.unpublished")
      )
    end
  end
end
