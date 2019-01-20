# frozen_string_literal: true

module RootResourceDecoratorCommon
  extend ActiveSupport::Concern

  included do
    def edit_db_root_resource_path
      case self.class.name
      when "Work" then edit_db_work_path(self)
      when "Character" then edit_db_character_path(self)
      when "Series" then edit_db_series_path(self)
      end
    end
  end
end
