# frozen_string_literal: true

module Api
  module Internal
    class WorksController < Api::Internal::ApplicationController
      def index
        q = params[:q]
        @works = if q
          Work.where("title ILIKE ?", "%#{q}%").published
        else
          Work.none
        end
      end

      def show
        @work = Work.published.find(params[:id])
      end
    end
  end
end
