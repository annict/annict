# frozen_string_literal: true

module DbActivityMethods
  extend ActiveSupport::Concern

  included do
    attr_accessor :user

    def save_and_create_activity!
      return false unless valid?

      action = format_action
      old_attrs = to_attributes
      ActiveRecord::Base.transaction do
        save!(validate: false)
        create_activity!(action, old_attrs)
      end
    end
  end

  def root_resource
    case self.class.name
    when "Work", "Person", "Organization", "Character", "Series"
      self
    when "Episode", "Slot", "Cast", "Staff"
      work
    when "SeriesWork"
      series
    end
  end

  private

  def create_activity!(action, old_attrs)
    new_attrs = to_attributes
    return if old_attrs == new_attrs

    DbActivity.create! do |a|
      a.user = @user
      a.root_resource = root_resource
      a.trackable = self
      a.action = action
      a.parameters = if old_attrs.blank?
        {new: new_attrs}
      else
        {old: old_attrs, new: new_attrs}
      end
    end
  end

  def to_attributes
    return nil if new_record?

    self.class.find(id).attributes
  end

  def format_action
    "#{self.class.table_name}.#{new_record? ? "create" : "update"}"
  end
end
