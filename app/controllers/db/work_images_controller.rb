# frozen_string_literal: true

module Db
  class WorkImagesController < Db::ApplicationController
    permits :attachment, :asin, :copyright

    before_action :authenticate_user!
    before_action :load_work, only: %i(show create update destroy)
    before_action :load_image, only: %i(update destroy)

    def show
      @image = @work.work_image.presence || @work.build_work_image
    end

    def create(work_image)
      @image = @work.build_work_image(work_image)
      authorize @image, :create?
      @image.user = current_user

      if @image.save
        @work.purge
        flash[:notice] = t "messages.work_images.saved"
        redirect_to db_work_image_path(@work)
      else
        render :show
      end
    end

    def update(work_image)
      authorize @image, :update?

      @image.attributes = work_image
      @image.user = current_user

      if @image.save
        @work.purge
        flash[:notice] = t "messages.work_images.saved"
        redirect_to db_work_image_path(@work)
      else
        render :show
      end
    end

    def destroy
      authorize @item, :destroy?
      @item.destroy
      @work.purge
      redirect_to db_work_image_path(@work), notice: t("messages.work_images.deleted")
    end

    private

    def load_image
      @image = @work.work_image
      raise ActiveRecord::RecordNotFound if @image.blank?
    end
  end
end
