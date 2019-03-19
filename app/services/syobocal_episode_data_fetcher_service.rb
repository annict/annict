# frozen_string_literal: true

class SyobocalEpisodeDataFetcherService
  def self.execute!(now: Time.current)
    new.execute!(now: now)
  end

  def execute!(now:)
    client = SyoboiCalendar::Client.new
    episodes = Episode.
      published.
      where(title: [nil, ""]).
      where.not(raw_number: nil).
      after(now - 7.days).
      joins(:work, :programs).
      merge(Work.where.not(sc_tid: nil)).
      merge(Program.where.not(program_detail_id: nil)).
      distinct
    works = Work.published.where(id: episodes.pluck(:work_id).uniq)
    titles = client.list_titles(title_id: works.pluck(:sc_tid))

    episodes.each do |e|
      next if e.raw_number.to_i != e.raw_number

      title = titles.find { |t| t.id == e.work.sc_tid }
      next unless title
      next unless title.sub_titles

      sub_titles = title.sub_titles.split("\n").map do |st|
        ary = st.split("*") # `"*01*アンダーワールド"` => `["", "01", "アンダーワールド"]`
        [ary[1].to_i.to_s, ary[2]]
      end.to_h
      next if sub_titles.empty?

      episode_title = sub_titles[e.raw_number.to_i.to_s]
      e.update_column(:title, episode_title) if episode_title.present?
    end
  end
end
