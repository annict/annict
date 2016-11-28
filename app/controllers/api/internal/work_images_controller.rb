# frozen_string_literal: true

module Api
  module Internal
    class WorkImagesController < Api::Internal::ApplicationController
      before_action :authenticate_user!
      before_action :load_work

      def create
        @image = @work.work_images.new(attachment: params[:file])
        @image.user = current_user

        if @image.save
          flash[:notice] = t "resources.work_image.uploaded"
          head 201
        else
          render status: 400, json: { message: "Error" }
        end
      end
    end
  end
end
