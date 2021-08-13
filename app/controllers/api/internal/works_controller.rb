# frozen_string_literal: true

module Api
  module Internal
    class WorksController < Api::Internal::ApplicationController
      def index
        q = params[:q]
        @works = if q
          Work.where("title ILIKE ?", "%#{q}%").only_kept
        else
          Work.none
        end
      end
    end
  end
end
