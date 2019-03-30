# frozen_string_literal: true

module Db
  class WorkImagesController < Db::ApplicationController
    before_action :authenticate_user!
    before_action :load_work, only: %i(show create update destroy)
    before_action :load_image, only: %i(update destroy)

    def show
      @image = @work.work_image.presence || @work.build_work_image
    end

    def create
      @image = @work.build_work_image(work_image_params)
      authorize @image, :create?
      @image.user = current_user

      if @image.save
        flash[:notice] = t "messages.work_images.saved"
        redirect_to db_work_image_path(@work)
      else
        render :show
      end
    end

    def update
      authorize @image, :update?

      @image.attributes = work_image_params
      @image.user = current_user

      if @image.save
        flash[:notice] = t "messages.work_images.saved"
        redirect_to db_work_image_path(@work)
      else
        render :show
      end
    end

    def destroy
      authorize @item, :destroy?
      @item.destroy
      redirect_to db_work_image_path(@work), notice: t("messages.work_images.deleted")
    end

    private

    def load_image
      @image = @work.work_image
      raise ActiveRecord::RecordNotFound if @image.blank?
    end

    def work_image_params
      params.require(:work_image).permit(:attachment, :asin, :copyright)
    end
  end
end
