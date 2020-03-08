# frozen_string_literal: true

module DB
  class VodTitlesController < DB::ApplicationController
    before_action :authenticate_user!, only: %i(new create edit update destroy)

    def index
      @vod_titles = VodTitle.order(id: :desc).page(params[:page])
    end

    def hide
      vod_title = VodTitle.find(params[:id])
      authorize vod_title, :hide?

      vod_title.soft_delete

      flash[:notice] = t("messages._common.unpublished")
      redirect_back fallback_location: db_vod_titles_path
    end
  end
end
