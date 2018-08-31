

new Vue({
    el: '#home',
    data: {
        photos: []
    },
    delimiters: ['${', '}'],
    mounted() {
        this.getPhotos();
      },
    methods:{
        getPhotos() {
            axios
            .get('/photos')
            .then(response => (this.photos = response.data))
            .catch(error => console.log(error))
        }
    }
});