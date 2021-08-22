# frozen_string_literal: true

# == Schema Information
#
# Table name: series_works
#
#  id             :bigint           not null, primary key
#  aasm_state     :string           default("published"), not null
#  deleted_at     :datetime
#  summary        :string           default(""), not null
#  summary_en     :string           default(""), not null
#  unpublished_at :datetime
#  created_at     :datetime         not null
#  updated_at     :datetime         not null
#  series_id      :bigint           not null
#  work_id        :bigint           not null
#
# Indexes
#
#  index_series_works_on_deleted_at             (deleted_at)
#  index_series_works_on_series_id              (series_id)
#  index_series_works_on_series_id_and_work_id  (series_id,work_id) UNIQUE
#  index_series_works_on_unpublished_at         (unpublished_at)
#  index_series_works_on_work_id                (work_id)
#
# Foreign Keys
#
#  fk_rails_...  (series_id => series.id)
#  fk_rails_...  (work_id => works.id)
#
SeriesAnime = SeriesWork
