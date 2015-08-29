describe "MultipleEpisodesFormatter" do
  describe "#to_episode_hash" do
    context "通常の話数、サブタイトルのとき" do
      let(:episode) { DraftMultipleEpisode.new(body: "#1,aaa") }

      it "カンマで区切って値を返すこと" do
        expect(episode.to_episode_hash).to eq [{ number: "#1", title: "aaa" }]
      end
    end

    context "サブタイトルにダブルクォートが入っているとき" do
      let(:episode) { DraftMultipleEpisode.new(body: '#1,a"a"a') }

      it "カンマで区切って値を返すこと" do
        expect(episode.to_episode_hash).to eq [{ number: "#1", title: 'a"a"a' }]
      end
    end
  end
end
