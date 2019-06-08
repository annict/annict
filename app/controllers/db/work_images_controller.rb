# frozen_string_literal: true

module Db
  class WorkImagesController < Db::ApplicationController
    before_action :authenticate_user!

    def show
      @work = Work.find(params[:work_id])
      @image = @work.work_image.presence || @work.build_work_image
    end

    def create
      @work = Work.find(params[:work_id])
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
      @work = Work.find(params[:work_id])
      @image = WorkImage.find_by!(work_id: @work.id)
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
      @work = Work.find(params[:work_id])
      @image = WorkImage.find_by!(work_id: @work.id)
      authorize @item, :destroy?
      @item.destroy
      redirect_to db_work_image_path(@work), notice: t("messages.work_images.deleted")
    end

    private

    def work_image_params
      params.require(:work_image).permit(:image, :asin, :copyright)
    end
  end
end
