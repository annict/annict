# frozen_string_literal: true

class CharacterImageReportsController < ApplicationController
  before_action :authenticate_user!, only: %i(create)

  def create(character_image_id)
    @image = CharacterImage.find(character_image_id)
    @report = Report.new do |r|
      r.user = current_user
      r.root_resource = @image.character
      r.resource = @image
    end

    if @report.save
      ReportMailer.notify(@report.id).deliver_later
      flash[:notice] = t "messages.character_image_reports.reported"
      redirect_to character_images_path(@image.character)
    else
      flash[:alert] = t "messages.character_image_reports.failed"
      redirect_to :back
    end
  end
end
