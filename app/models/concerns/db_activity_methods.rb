# frozen_string_literal: true

module DbActivityMethods
  extend ActiveSupport::Concern

  included do
    def save_and_create_activity!(user)
      return false unless valid?

      @user = user

      old_attrs = _to_attributes
      ActiveRecord::Base.transaction do
        save!(validate: false)
        create_activity!(old_attrs)
      end
    end
  end

  private

  def create_activity!(old_attrs)
    new_attrs = _to_attributes
    return if old_attrs == new_attrs

    DbActivity.create! do |a|
      a.user = @user
      a.root_resource = _root_resource
      a.trackable = _trackable_resource
      a.action = _action
      a.parameters = if old_attrs.blank?
        { new: new_attrs }
      else
        { old: old_attrs, new: new_attrs }
      end
    end
  end

  def _to_attributes
    case self.class.name
    when "Work"
      return nil if new_record?
      self.class.find(id).attributes
    else
      to_attributes
    end
  end

  def _root_resource
    case self.class.name
    when "Work" then self
    else
      root_resource
    end
  end

  def _trackable_resource
    case self.class.name
    when "Work" then self
    else
      trackable_resource
    end
  end

  def _action
    case self.class.name
    when "Work"
      action = new_record? ? "create" : "update"
      "#{self.class.table_name}.#{action}"
    else
      table_name = trackable_resource.class.table_name
      action = new_record ? "create" : "update"
      "#{table_name}.#{action}"
    end
  end
end
