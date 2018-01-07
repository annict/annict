# frozen_string_literal: true

module Db
  class VodTitlesController < Db::ApplicationController
    before_action :authenticate_user!, only: %i(new create edit update destroy)

    def index(page: nil)
      @vod_titles = VodTitle.order(id: :desc).page(page)
    end

    def hide(id)
      vod_title = VodTitle.find(id)
      authorize vod_title, :hide?

      vod_title.hide!

      flash[:notice] = t("messages._common.unpublished")
      redirect_back fallback_location: db_vod_titles_path
    end
  end
end
