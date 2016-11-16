# frozen_string_literal: true

class WorkImagesController < ApplicationController
  before_action :load_work, only: %i(index show)

  def index
    @images = @work.work_images.published
  end

  def new
    @image = @work.work_images.new
  end
end
