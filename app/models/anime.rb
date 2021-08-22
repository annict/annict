# frozen_string_literal: true

# == Schema Information
#
# Table name: works
#
#  id                           :bigint           not null, primary key
#  aasm_state                   :string           default("published"), not null
#  deleted_at                   :datetime
#  ended_on                     :date
#  episodes_count               :integer          default(0), not null
#  facebook_og_image_url        :string           default(""), not null
#  manual_episodes_count        :integer
#  media                        :integer          not null
#  no_episodes                  :boolean          default(FALSE), not null
#  official_site_url            :string(510)      default(""), not null
#  official_site_url_en         :string           default(""), not null
#  ratings_count                :integer          default(0), not null
#  recommended_image_url        :string           default(""), not null
#  records_count                :integer          default(0), not null
#  released_at                  :date
#  released_at_about            :string
#  satisfaction_rate            :float
#  sc_tid                       :integer
#  score                        :float
#  season_name                  :integer
#  season_year                  :integer
#  start_episode_raw_number     :float            default(1.0), not null
#  started_on                   :date
#  synopsis                     :text             default(""), not null
#  synopsis_en                  :text             default(""), not null
#  synopsis_source              :string           default(""), not null
#  synopsis_source_en           :string           default(""), not null
#  title                        :string(510)      not null
#  title_alter                  :string           default(""), not null
#  title_alter_en               :string           default(""), not null
#  title_en                     :string           default(""), not null
#  title_kana                   :string           default(""), not null
#  title_ro                     :string           default(""), not null
#  twitter_hashtag              :string(510)
#  twitter_image_url            :string           default(""), not null
#  twitter_username             :string(510)
#  unpublished_at               :datetime
#  watchers_count               :integer          default(0), not null
#  wikipedia_url                :string(510)      default(""), not null
#  wikipedia_url_en             :string           default(""), not null
#  work_records_count           :integer          default(0), not null
#  work_records_with_body_count :integer          default(0), not null
#  created_at                   :datetime
#  updated_at                   :datetime
#  key_pv_id                    :bigint
#  mal_anime_id                 :integer
#  number_format_id             :bigint
#  season_id                    :bigint
#
# Indexes
#
#  index_works_on_aasm_state                           (aasm_state)
#  index_works_on_deleted_at                           (deleted_at)
#  index_works_on_key_pv_id                            (key_pv_id)
#  index_works_on_number_format_id                     (number_format_id)
#  index_works_on_ratings_count                        (ratings_count)
#  index_works_on_satisfaction_rate                    (satisfaction_rate)
#  index_works_on_satisfaction_rate_and_ratings_count  (satisfaction_rate,ratings_count)
#  index_works_on_score                                (score)
#  index_works_on_season_year                          (season_year)
#  index_works_on_season_year_and_season_name          (season_year,season_name)
#  index_works_on_unpublished_at                       (unpublished_at)
#  works_season_id_idx                                 (season_id)
#
# Foreign Keys
#
#  fk_rails_...        (key_pv_id => trailers.id)
#  fk_rails_...        (number_format_id => number_formats.id)
#  works_season_id_fk  (season_id => seasons.id) ON DELETE => cascade
#
Anime = Work
