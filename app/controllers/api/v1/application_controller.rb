# frozen_string_literal: true

module Api
  module V1
    class ApplicationController < ActionController::Base
      before_action :doorkeeper_authorize!

      private

      def prepare_params!
        class_name = self.class.name
        class_name = class_name.sub("Controller", "#{params[:action].classify}Params")
        @params = Object.const_get(class_name).new(params)
        return response_params_error unless @params.valid?
      end

      def response_params_error
        data = {
          errors: [{
            type: "invalid_params",
            message: @params.errors.full_messages.first,
            url: "http://example.com/docs/api/validations"
          }]
        }

        render json: data, status: 400
      end
    end
  end
end
