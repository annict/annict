# frozen_string_literal: true

module Api
  module Internal
    class WorksController < Api::Internal::ApplicationController
      def index(q: nil)
        @works = if q.present?
          Work.where("title ILIKE ?", "%#{q}%").published
        else
          Work.none
        end
      end
    end
  end
end
