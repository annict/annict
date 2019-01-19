# frozen_string_literal: true

module RootResourceDecoratorCommon
  extend ActiveSupport::Concern

  included do
    def edit_db_root_resource_path
      case model.class.name
      when "Work" then edit_db_work_path(model)
      when "Character" then edit_db_character_path(model)
      when "Series" then edit_db_series_path(model)
      end
    end
  end
end
