# == Schema Information
#
# Table name: works
#
#  id                :integer          not null, primary key
#  title             :string           not null
#  media             :integer          not null
#  official_site_url :string           default(""), not null
#  wikipedia_url     :string           default(""), not null
#  released_at       :date
#  created_at        :datetime
#  updated_at        :datetime
#  episodes_count    :integer          default(0), not null
#  season_id         :integer
#  twitter_username  :string
#  twitter_hashtag   :string
#  watchers_count    :integer          default(0), not null
#  sc_tid            :integer
#  released_at_about :string
#  aasm_state        :string           default("published"), not null
#  number_format_id  :integer
#  title_kana        :string           default(""), not null
#
# Indexes
#
#  index_works_on_aasm_state        (aasm_state)
#  index_works_on_episodes_count    (episodes_count)
#  index_works_on_media             (media)
#  index_works_on_number_format_id  (number_format_id)
#  index_works_on_released_at       (released_at)
#  index_works_on_sc_tid            (sc_tid) UNIQUE
#  index_works_on_watchers_count    (watchers_count)
#

module WorksHelper
  def shirobako_color(round)
    return "shirobako-#{round}" if round >= 1 && round <= 6
    ''
  end
end
