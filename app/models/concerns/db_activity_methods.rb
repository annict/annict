# frozen_string_literal: true

module DbActivityMethods
  extend ActiveSupport::Concern

  included do
    def save_and_create_db_activity(user, action)
      return false unless valid?

      case action
      when "multiple_episodes.create"
        ActiveRecord::Base.transaction do
          Episode.create_from_multiple_episodes(work, to_episode_hash)
          data = {
            user: user,
            root_resource: work,
            old_attrs: nil,
            new_attrs: to_episode_hash,
            action: action
          }
          create_db_activity(data)
        end
      else
        ActiveRecord::Base.transaction do
          old_attrs = to_attributes
          save(validate: false)
          new_attrs = to_attributes
          data = {
            user: user,
            root_resource: root_resource,
            old_attrs: old_attrs,
            new_attrs: new_attrs,
            action: action
          }
          create_db_activity(data)
        end
      end
    end
  end

  private

  def create_db_activity(data)
    if data[:old_attrs] != data[:new_attrs]
      DbActivity.create do |a|
        a.user = data[:user]
        a.root_resource = data[:root_resource]
        a.trackable = self
        a.action = data[:action]
        a.parameters = if data[:old_attrs].blank?
          { new: data[:new_attrs] }
        else
          { old: data[:old_attrs], new: data[:new_attrs] }
        end
      end
    end
  end

  def to_attributes
    return nil if new_record?

    self.class.find(id).attributes
  end

  def root_resource
    case self.class.name
    when "Work" then self
    end
  end
end
