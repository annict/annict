describe "TwitterService" do
  describe "#tweet_body" do
    context "話数表記にシャープが含まれるとき" do
      let(:work) { create(:work, title: "2人はプリキュア") }
      let(:episode) { create(:episode, work: work, number: "#1", title: "Yes! プリキュア") }

      context "コメント付きの記録のとき" do
        let(:checkin) { create(:checkin, episode: episode, comment: "良かった") }
        let(:twitter) { TwitterService.new(checkin.user) }

        it "ツイート本文が成形されること" do
          tweet = twitter.send(:tweet_body, checkin)
          expect(tweet).to eq "良かった／2人はプリキュア #1 を見ました localhost:3000/r/tw/xxxxx #precure"
        end
      end

      context "コメント無しの記録のとき" do
        let(:checkin) { create(:checkin, comment: "", episode: episode) }
        let(:twitter) { TwitterService.new(checkin.user) }

        it "ツイート本文が成形されること" do
          tweet = twitter.send(:tweet_body, checkin)
          expect(tweet).to eq "2人はプリキュア #1 「Yes! プリキュア」を見ました localhost:3000/r/tw/xxxxx #precure"
        end
      end
    end

    context "話数表記にシャープが含まれないとき" do
      let(:work) { create(:work, title: "2人はプリキュア") }
      let(:episode) { create(:episode, work: work, number: "第1話", title: "Yes! プリキュア") }

      context "コメント付きの記録のとき" do
        let(:checkin) { create(:checkin, episode: episode, comment: "良かった") }
        let(:twitter) { TwitterService.new(checkin.user) }

        it "ツイート本文が成形されること" do
          tweet = twitter.send(:tweet_body, checkin)
          expect(tweet).to eq "良かった／2人はプリキュア 第1話を見ました localhost:3000/r/tw/xxxxx #precure"
        end
      end

      context "コメント無しの記録のとき" do
        let(:checkin) { create(:checkin, comment: "", episode: episode) }
        let(:twitter) { TwitterService.new(checkin.user) }

        it "ツイート本文が成形されること" do
          tweet = twitter.send(:tweet_body, checkin)
          expect(tweet).to eq "2人はプリキュア 第1話「Yes! プリキュア」を見ました localhost:3000/r/tw/xxxxx #precure"
        end
      end
    end

    describe "ハッシュタグのチェック" do
      let(:work) { create(:work, title: "2人はプリキュア") }
      let(:episode) { create(:episode, work: work, number: "第1話", title: "Yes! プリキュア") }
      let(:checkin) { create(:checkin, episode: episode) }

      context "ハッシュタグが設定されているとき" do
        before do
          work.update_column(:twitter_hashtag, "precure")
        end

        it "ツイート本文が成形されること" do
          twitter = TwitterService.new(checkin.user)
          tweet = twitter.send(:tweet_body, checkin)
          expect(tweet).to eq "おもしろかった／2人はプリキュア 第1話を見ました localhost:3000/r/tw/xxxxx #precure"
        end
      end

      context "ハッシュタグが設定されていないとき" do
        before do
          work.update_column(:twitter_hashtag, nil)
        end

        it "ツイート本文が成形されること" do
          twitter = TwitterService.new(checkin.user)
          tweet = twitter.send(:tweet_body, checkin)
          expect(tweet).to eq "おもしろかった／2人はプリキュア 第1話を見ました localhost:3000/r/tw/xxxxx"
        end
      end
    end
  end
end
