module DbActivityMethods
  extend ActiveSupport::Concern

  included do
    def save_and_create_db_activity(user, action)
      return false unless valid?

      case action
      when "multiple_episodes.create"
        ActiveRecord::Base.transaction do
          Episode.create_from_multiple_episodes(work, to_episode_hash)
          create_db_activity(user, nil, to_episode_hash, action, trackable: work)
        end
      else
        ActiveRecord::Base.transaction do
          old_attributes = to_attributes
          save(validate: false)
          new_attributes = to_attributes
          create_db_activity(user, old_attributes, new_attributes, action)
        end
      end
    end
  end

  private

  def create_db_activity(user, old_attributes, new_attributes, action, trackable: self)
    if old_attributes != new_attributes
      DbActivity.create do |dba|
        dba.user = user
        dba.trackable = trackable
        dba.action = action
        dba.parameters = if old_attributes.blank?
          { new: new_attributes }
        else
          { old: old_attributes, new: new_attributes }
        end
      end
    end
  end

  def to_attributes
    return nil if new_record?

    self.class.find(id).attributes
  end
end
