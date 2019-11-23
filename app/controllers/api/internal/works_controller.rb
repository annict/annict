# frozen_string_literal: true

module Api
  module Internal
    class WorksController < Api::Internal::ApplicationController
      def index
        q = params[:q]
        @works = if q
          Work.where("title ILIKE ?", "%#{q}%").without_deleted
        else
          Work.none
        end
      end

      def show
        @work = Work.without_deleted.find(params[:id])
      end
    end
  end
end
