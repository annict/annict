# frozen_string_literal: true

module Fragment
  class ActivityItemsController < Fragment::ApplicationController
    def index
      set_page_category params[:page_category]

      @activity_group = ActivityGroup.find(params[:activity_group_id])
      @work_ids = @activity_group.activity_items.map(&:work_id).uniq
    end
  end
end
