
new Vue({
    el: '#vote',
    data: {
        photos: [],
    },
    delimiters: ['${', '}'],
    mounted() {
        this.getPhotos();
      },
    methods: {
        getPhotos() {
            axios
            .get('/photos/stack')
            .then(response => (this.photos = response.data))
            .catch(error => console.log(error))
        },

        submitVote(photo_id, vote_choice, index) {
            axios
            .post('/photos/vote', {photo: photo_id, vote:vote_choice})
            .then(
                this.removeRow(index)
                
            )
            .catch(error => console.log(error))
        },

        removeRow(index) {
            this.photos.splice(index, 1);
        }
    }
});