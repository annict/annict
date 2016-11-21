# frozen_string_literal: true

class WorkImagesController < ApplicationController
  before_action :authenticate_user!, only: %i(destroy)
  before_action :load_work, only: %i(index destroy)

  def index
    @images = @work.work_images.published
  end

  def destroy(id)
    @image = @work.work_images.find(id)
    @image.destroy
    redirect_to work_images_path(@work), notice: t("resources.work_image.deleted")
  end
end
