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
          old_diffable_hash = new_record? ? nil : self.class.find(id).to_diffable_hash
          save(validate: false)
          create_db_activity(user, old_diffable_hash, to_diffable_hash, action)
        end
      end
    end
  end

  private

  def create_db_activity(user, old_diffable_hash, new_diffable_hash, action, trackable: self)
    DbActivity.create do |dba|
      dba.user = user
      dba.trackable = trackable
      dba.action = action
      dba.parameters = if old_diffable_hash.blank?
        { new: new_diffable_hash }
      else
        { old: old_diffable_hash, new: new_diffable_hash }
      end
    end
  end
end
