# frozen_string_literal: true

describe Destroyers::RecordDestroyer, type: :model do
  let!(:user) { create :registered_user }
  let!(:work) { create :work }

  context "エピソードへの記録を削除するとき" do
    let!(:episode) { create :episode, work: work }
    let!(:record) { create :record, :on_episode, user: user, work: work, episode: episode }

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
    let!(:record) { create :record, :on_work, user: user, work: work }

    it "削除できること" do
      # Destroyerを呼ぶ前なので、各レコードは1件のはず
      expect(Record.count).to eq 1
      expect(WorkRecord.count).to eq 1

      Destroyers::RecordDestroyer.new(record: record).call

      expect(Record.count).to eq 0
      expect(WorkRecord.count).to eq 0
    end
  end
end
