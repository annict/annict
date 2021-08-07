# frozen_string_literal: true

describe Destroyers::RecordDestroyer, type: :model do
  let!(:user) { create :registered_user }
  let!(:anime) { create :anime }
  let!(:record) { create :record, user: user, anime: anime }

  context "エピソードへの記録を削除するとき" do
    let!(:episode) { create :episode, anime: anime }
    let!(:episode_record) { create :episode_record, user: user, anime: anime, episode: episode, record: record }

    it "削除できること" do
      # Destroyerを呼ぶ前なので、各レコードは1件のはず
      expect(Record.count).to eq 1
      expect(EpisodeRecord.count).to eq 1

      Destroyers::RecordDestroyer.new(record: record).call

      expect(Record.count).to eq 0
      expect(EpisodeRecord.count).to eq 0
    end
  end

  context "アニメへの記録を削除するとき" do
    let!(:anime_record) { create :anime_record, user: user, anime: anime, record: record }

    it "削除できること" do
      # Destroyerを呼ぶ前なので、各レコードは1件のはず
      expect(Record.count).to eq 1
      expect(AnimeRecord.count).to eq 1

      Destroyers::RecordDestroyer.new(record: record).call

      expect(Record.count).to eq 0
      expect(AnimeRecord.count).to eq 0
    end
  end
end
