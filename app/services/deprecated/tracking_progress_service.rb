# frozen_string_literal: true

class Deprecated::TrackingProgressService
  attr_reader :user, :work, :episode_ids, :checked_episode_ids, :all_records_count

  def initialize(user, work)
    @user = user
    @work = work
    @episode_ids = work.episodes.only_kept.pluck(:id)
    @checked_episode_ids = user.records.where(work: work).pluck(:episode_id)
    @all_records_count = get_all_records_count(checked_episode_ids)
  end

  # 指定した作品の視聴が何周目かを返す
  def episodes_round
    count = episode_ids.count * all_records_count

    if checked_episode_ids.count > count
      all_records_count + 1
    elsif checked_episode_ids.count == count
      all_records_count
    else
      0
    end
  end

  def halfway_checked_count
    ids = checked_episode_ids

    if all_checked? && over_tracking?
      all_records_count.times do
        ids = remove_checked_episode_id(ids)
      end

      return ids.uniq.count
    end

    (episode_ids & ids).count
  end

  def ratio
    begin
      halfway_checked_count / work.episodes.only_kept.count.to_f
    rescue
      1
    end * 100
  end

  private

  # 何周見たかを返す
  def get_all_records_count(checked_episode_ids, count = 0)
    unchecked_episode_ids = episode_ids - checked_episode_ids

    return count if unchecked_episode_ids.present?

    if checked_episode_ids.present?
      ids = remove_checked_episode_id(checked_episode_ids)
      get_all_records_count(ids, count + 1)
    else
      count
    end
  end

  # 全てのエピソードを記録しているかどうか
  def all_checked?
    episode_ids.count == (episode_ids & checked_episode_ids).count
  end

  # 全てのエピソードを記録した上で、さらに何本かのエピソードを記録しているかどうか
  def over_tracking?
    checked_episode_ids.count > episode_ids.count * all_records_count
  end

  def remove_checked_episode_id(checked_episode_ids)
    ids = checked_episode_ids.dup

    episode_ids.each do |eid|
      i = ids.index(eid)
      ids.delete_at(i) if i.present?
    end

    ids
  end
end
