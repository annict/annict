# typed: false
# frozen_string_literal: true

module RootResourceDecoratorCommon
  extend ActiveSupport::Concern

  included do
    def db_edit_root_resource_path
      case self.class.name
      when "Work" then db_edit_work_path(self)
      when "Character" then db_edit_character_path(self)
      when "Series" then db_edit_series_path(self)
      end
    end
  end
end
