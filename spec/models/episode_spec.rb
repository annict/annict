describe Episode do
  describe '#create_nicoch_program' do
    let!(:channel) { create(:channel, name: 'ニコニコチャンネル') }
    let(:work) { create(:work, nicoch_started_at: Date.today) }
    let(:episode) { create(:episode, work: work) }

    it 'ニコニコチャンネルの番組予定が保存されること' do
      expect(episode.programs.first.channel).to eq channel
    end
  end
end
