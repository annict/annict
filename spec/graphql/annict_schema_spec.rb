# frozen_string_literal: true

describe AnnictSchema do
  describe "node" do
    before do
      @work = create(:work)
      @id = GraphQL::Schema::UniqueWithinType.encode(@work.class.name, @work.id)
      query_string = <<~GRAPHQL
        query {
          node(id: "#{@id}") {
            id
            ... on Work {
              annictId
              title
            }
          }
        }
      GRAPHQL
      @res = AnnictSchema.execute(query_string)
    end

    it "fetches resource" do
      expected = {
        data: {
          node: {
            id: @id,
            annictId: @work.id,
            title: @work.title
          }
        }
      }
      expect(@res.to_h.deep_stringify_keys).to include(expected.deep_stringify_keys)
    end
  end

  describe "nodes" do
    before do
      @work1 = create(:work)
      @work2 = create(:work)
      @id1 = GraphQL::Schema::UniqueWithinType.encode(@work1.class.name, @work1.id)
      @id2 = GraphQL::Schema::UniqueWithinType.encode(@work2.class.name, @work2.id)
      query_string = <<~GRAPHQL
        query {
          nodes(ids: ["#{@id1}", "#{@id2}"]) {
            id
            ... on Work {
              annictId
              title
            }
          }
        }
      GRAPHQL
      @res = AnnictSchema.execute(query_string)
    end

    it "fetches resources" do
      expected = {
        data: {
          nodes: [
            {
              id: @id1,
              annictId: @work1.id,
              title: @work1.title
            },
            {
              id: @id2,
              annictId: @work2.id,
              title: @work2.title
            }
          ]
        }
      }
      expect(@res.to_h.deep_stringify_keys).to include(expected.deep_stringify_keys)
    end
  end

  describe "user" do
    before do
      @user = create(:user)
      query_string = <<~GRAPHQL
        query {
          user(username: "#{@user.username}") {
            annictId
            username
          }
        }
      GRAPHQL
      @res = AnnictSchema.execute(query_string)
    end

    it "fetches user" do
      expected = {
        data: {
          user: {
            annictId: @user.id,
            username: @user.username
          }
        }
      }
      expect(@res.to_h.deep_stringify_keys).to include(expected.deep_stringify_keys)
    end
  end

  describe "viewer" do
    before do
      @user = create(:user)
      query_string = <<~GRAPHQL
        query {
          viewer {
            annictId
            username
          }
        }
      GRAPHQL
      @res = AnnictSchema.execute(query_string, context: { viewer: @user })
    end

    it "fetches user" do
      expected = {
        data: {
          viewer: {
            annictId: @user.id,
            username: @user.username
          }
        }
      }
      expect(@res.to_h.deep_stringify_keys).to include(expected.deep_stringify_keys)
    end
  end

  describe "searchWorks" do
    context "with `annictIds` argument" do
      before do
        @work = create(:work)
        query_string = <<~GRAPHQL
          query {
            searchWorks(annictIds: [#{@work.id}]) {
              edges {
                node {
                  annictId
                  title
                }
              }
            }
          }
        GRAPHQL
        @res = AnnictSchema.execute(query_string)
      end

      it "fetches works" do
        expected = {
          data: {
            searchWorks: {
              edges: [
                {
                  node: {
                    annictId: @work.id,
                    title: @work.title
                  }
                }
              ]
            }
          }
        }
        expect(@res.to_h.deep_stringify_keys).to include(expected.deep_stringify_keys)
      end
    end

    context "with `seasons` argument" do
      before do
        @work = create(:work, :with_current_season)
        query_string = <<~GRAPHQL
          query {
            searchWorks(seasons: ["#{@work.season.slug}"]) {
              edges {
                node {
                  annictId
                  title
                }
              }
            }
          }
        GRAPHQL
        @res = AnnictSchema.execute(query_string)
      end

      it "fetches works" do
        expected = {
          data: {
            searchWorks: {
              edges: [
                {
                  node: {
                    annictId: @work.id,
                    title: @work.title
                  }
                }
              ]
            }
          }
        }
        expect(@res.to_h.deep_stringify_keys).to include(expected.deep_stringify_keys)
      end
    end

    context "with `titles` argument" do
      before do
        @work = create(:work)
        query_string = <<~GRAPHQL
          query {
            searchWorks(titles: ["#{@work.title}"]) {
              edges {
                node {
                  annictId
                  title
                }
              }
            }
          }
        GRAPHQL
        @res = AnnictSchema.execute(query_string)
      end

      it "fetches works" do
        expected = {
          data: {
            searchWorks: {
              edges: [
                {
                  node: {
                    annictId: @work.id,
                    title: @work.title
                  }
                }
              ]
            }
          }
        }
        expect(@res.to_h.deep_stringify_keys).to include(expected.deep_stringify_keys)
      end
    end

    context "with `orderBy` argument" do
      before do
        @work1 = create(:work, :with_current_season)
        @work2 = create(:work, :with_next_season)

        query_string = <<~GRAPHQL
          query {
            searchWorks(orderBy: { field: SEASON, direction: DESC }) {
              edges {
                node {
                  annictId
                  title
                }
              }
            }
          }
        GRAPHQL
        @res = AnnictSchema.execute(query_string)
      end

      it "fetches ordered works" do
        expected = {
          data: {
            searchWorks: {
              edges: [
                {
                  node: {
                    annictId: @work2.id,
                    title: @work2.title
                  }
                },
                {
                  node: {
                    annictId: @work1.id,
                    title: @work1.title
                  }
                }
              ]
            }
          }
        }
        expect(@res.to_h.deep_stringify_keys).to include(expected.deep_stringify_keys)
      end
    end
  end
end
