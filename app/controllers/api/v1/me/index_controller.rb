# frozen_string_literal: true

module Api
  module V1
    module Me
      class IndexController < Api::V1::ApplicationController
        before_action :prepare_params!, only: %i[show]

        def show
          @user = current_user
        end
      end
    end
  end
end
