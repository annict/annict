# frozen_string_literal: true

# == Schema Information
#
# Table name: work_records
#
#  id                     :bigint           not null, primary key
#  aasm_state             :string           default("published"), not null
#  body                   :text             not null
#  deleted_at             :datetime
#  impressions_count      :integer          default(0), not null
#  likes_count            :integer          default(0), not null
#  locale                 :string           default("other"), not null
#  modified_at            :datetime
#  rating_animation_state :string
#  rating_character_state :string
#  rating_music_state     :string
#  rating_overall_state   :string
#  rating_story_state     :string
#  title                  :string           default("")
#  created_at             :datetime         not null
#  updated_at             :datetime         not null
#  oauth_application_id   :bigint
#  record_id              :bigint           not null
#  user_id                :bigint           not null
#  work_id                :bigint           not null
#
# Indexes
#
#  index_work_records_on_deleted_at            (deleted_at)
#  index_work_records_on_locale                (locale)
#  index_work_records_on_oauth_application_id  (oauth_application_id)
#  index_work_records_on_record_id             (record_id) UNIQUE
#  index_work_records_on_user_id               (user_id)
#  index_work_records_on_work_id               (work_id)
#
# Foreign Keys
#
#  fk_rails_...  (oauth_application_id => oauth_applications.id)
#  fk_rails_...  (record_id => records.id)
#  fk_rails_...  (user_id => users.id)
#  fk_rails_...  (work_id => works.id)
#
AnimeRecord = WorkRecord
