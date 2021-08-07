# frozen_string_literal: true

module Db
  class WorkImagesController < Db::ApplicationController
    before_action :authenticate_user!

    def show
      @work = Anime.without_deleted.find(params[:work_id])
      @image = @work.anime_image.presence || @work.build_anime_image
    end

    def create
      @work = Anime.without_deleted.find(params[:work_id])
      @image = @work.build_anime_image(work_image_params)
      authorize @image
      @image.user = current_user

      return render(:show) unless @image.valid?

      @image.save

      redirect_to db_work_image_detail_path(@work), notice: t("messages.work_images.saved")
    end

    def update
      @work = Anime.without_deleted.find(params[:work_id])
      @image = AnimeImage.find_by!(work_id: @work.id)
      authorize @image

      @image.attributes = work_image_params
      @image.user = current_user

      return render(:show) unless @image.valid?

      @image.save

      redirect_to db_work_image_detail_path(@work), notice: t("messages.work_images.saved")
    end

    private

    def work_image_params
      params.require(:anime_image).permit(:image, :copyright)
    end
  end
end
